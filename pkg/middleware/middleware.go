package middleware

import (
	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
	"reflect"
)

//type Middleware func(httprouter.Handle) httprouter.Handle

type Middleware struct {
	Handler func(httprouter.Handle) httprouter.Handle
}

type MwStruct struct {
	Models *gorm.DB
}

type List struct {
	Middlewares map[string][]Middleware
}

func (m *List) Set(name string, middlewares []Middleware) *List {
	m.Middlewares[name] = middlewares

	return m
}

func (m *List) Chain(f httprouter.Handle, name ...string) httprouter.Handle {
	var middlewares []Middleware

	for _, n := range name {
		middlewares = append(middlewares, m.Middlewares[n]...)
	}

	middlewares = Unique(middlewares)

	// run original function
	return Chain(f, middlewares)
}

func (m *List) Get(name string) []Middleware {
	return m.Middlewares[name]
}

func (m *List) AppendFromCurrent(from string, to string) *List {
	m.Middlewares[to] = append(m.Middlewares[to], m.Middlewares[from]...)

	return m
}

func (m *List) PrependFromCurrent(from string, to string) *List {
	var tmp []Middleware
	tmp = append(tmp, m.Middlewares[from]...)
	tmp = append(tmp, m.Middlewares[to]...)
	m.Middlewares[to] = tmp

	tmp = nil

	return m
}

func (m *List) Append(to string, middlewares ...Middleware) *List {
	m.Middlewares[to] = append(m.Middlewares[to], middlewares...)

	return m
}

func (m *List) Prepend(to string, middlewares ...Middleware) *List {
	var tmp []Middleware
	tmp = append(tmp, middlewares...)
	tmp = append(tmp, m.Middlewares[to]...)
	m.Middlewares[to] = tmp

	tmp = nil

	return m
}

// Chain - chains all middleware functions right to left
// https://husobee.github.io/golang/http/middleware/2015/12/22/simple-middleware.html
func Chain(f httprouter.Handle, m []Middleware) httprouter.Handle {
	// if our chain is done, use the original handlerfunc
	if len(m) == 0 {
		return f
	}

	// otherwise run recursively over nested handlers
	return m[0].Handler(Chain(f, m[1:]))
}

func InitMiddlewareList(db *gorm.DB) *List {
	list := &List{
		Middlewares: make(map[string][]Middleware),
	}

	mw := &MwStruct{
		Models: db,
	}

	list = list.Set("default", []Middleware{
		{Handler: mw.LogRequest},
	}).Set("auth", []Middleware{
		{Handler: mw.Auth},
		{Handler: mw.LogRequest},
	})

	return list
}

// Unique removes duplicates from a slice of Middleware
func Unique(middlewares []Middleware) []Middleware {
	unique := make([]Middleware, 0, len(middlewares))

	found := false

	for _, v := range middlewares {
		found = false

		for _, u := range unique {
			// compare u.Handler and v.Handler
			if reflect.ValueOf(u.Handler).Pointer() == reflect.ValueOf(v.Handler).Pointer() {
				found = true
				break
			}
		}

		if !found {
			unique = append(unique, v)
		}
	}

	return unique
}
