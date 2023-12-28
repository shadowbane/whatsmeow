package rpc

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gomeow/cmd/common/dto"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"strings"
)

func GrpcSessionService() *SessionService {
	return &SessionService{}
}

type SessionService struct {
	dto.UnimplementedSessionServiceServer
	App *application.Application
}

func (s SessionService) Connect(_ context.Context, param *dto.SessionRequest) (*dto.SessionResponse, error) {
	var device models.Device

	s.App.Models.
		Where("code = ?", param.GetId()).
		Find(&device)

	if device.Code == "" {
		zap.S().Debugf("Device with ID %s not found", param.GetId())
		return nil, fmt.Errorf("device with id %s not found", param.GetId())
	}

	// if device already connected, return error
	if wmeow.ClientPointer[device.ID] != nil {
		zap.S().Debugf("Device already connected %v", device)
		return nil, errors.New("device already connected")
	}

	var subscribedEvents []string
	subscribedEvents = append(subscribedEvents, "All")
	device.Events = strings.Join(subscribedEvents, ",")

	result := s.App.Models.Save(&device)
	if result.Error != nil {
		zap.S().Debugf("Error updating device: %+v", result)
		return nil, fmt.Errorf("error updating device: %+v", result)
	}

	go func() {
		err := wmeow.StartClient(&device, s.App, device.JID.String, subscribedEvents)
		if err != nil {
			zap.S().Errorf("Error starting client: %+v", err)
		}
	}()

	return &dto.SessionResponse{
		Message: "device connected",
		Id:      param.GetId(),
		State:   dto.State_CONNECTED.String(),
	}, nil
}

func (s SessionService) Disconnect(_ context.Context, param *dto.SessionRequest) (*dto.SessionResponse, error) {
	var device models.Device

	s.App.Models.
		Where("code = ?", param.GetId()).
		Find(&device)

	if device.Code == "" {
		zap.S().Debugf("Device with ID %s not found", param.GetId())
		return nil, fmt.Errorf("device with id %s not found", param.GetId())
	}

	// if device already disconnected, return error
	if wmeow.ClientPointer[device.ID] == nil {
		zap.S().Debugf("Device already disconnected %v", device)
		return nil, errors.New("device already disconnected")
	}

	wmeow.KillChannel[device.ID] <- true

	return &dto.SessionResponse{
		Message: "device disconnected",
		Id:      param.GetId(),
		State:   dto.State_DISCONNECTED.String(),
	}, nil
}

func (s SessionService) Logout(_ context.Context, param *dto.SessionRequest) (*dto.SessionResponse, error) {
	var device models.Device

	s.App.Models.
		Where("code = ?", param.GetId()).
		Find(&device)

	if device.Code == "" {
		zap.S().Debugf("Device with ID %s not found", param.GetId())
		return nil, fmt.Errorf("device with id %s not found", param.GetId())
	}

	if device.JID.String == "" {
		zap.S().Debugf("Device with ID %s not logged in", param.GetId())
		return nil, fmt.Errorf("device with id %s not logged in", param.GetId())
	}

	client := wmeow.ClientPointer[device.ID].WAClient
	if client.IsConnected() && client.IsLoggedIn() {
		err := client.Logout()
		if err != nil {
			zap.S().Errorf("Error logging out: %+v", err)
			return nil, fmt.Errorf("error logging out: %+v", err)
		}

		wmeow.ClientPointer[device.ID].Logout()
		zap.S().Infof("Device %s with JID %s Logged Out", device.Name, device.JID.String)
	} else {
		if client.IsConnected() {
			zap.S().Infof("Device %s is not logged in. Doing logout anyway", device.Name)
			wmeow.KillChannel[device.ID] <- true
		} else {
			zap.S().Errorf("Device not connected")
			return nil, fmt.Errorf("device not connected")
		}
	}

	return &dto.SessionResponse{
		Message: "device logged out",
		Id:      param.GetId(),
		State:   dto.State_LOGGED_OUT.String(),
	}, nil
}
