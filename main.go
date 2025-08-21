package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"strings"

	"github.com/valyala/fasthttp"
)

type Owner struct {
	Name   string
	Street string
	Zip    string
	Place  string
}

type Content struct {
	Host        string
	Logo        []byte
	ContactMail string
	Owner       *Owner
}

type Data struct {
	Tmpl    *template.Template
	Content *Content
}

func logo() (data []byte, e error) {
	envLogo := os.Getenv("DOMAIN_LOGO")

	if strings.Contains(envLogo, "http") {

		if _, e = url.Parse(envLogo); e != nil {
			return
		}

		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		req.SetRequestURI(envLogo)

		defer func() {
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
		}()

		// get logo from env var DOMAIN_LOGO URI
		e = fasthttp.Do(req, resp)
		if e != nil {
			return
		}

		body := resp.Body()

		data = make([]byte, base64.StdEncoding.EncodedLen(len(body)))
		base64.StdEncoding.Encode(data, body)
	} else {
		return os.ReadFile("public/b64_logo_128x128")
	}

	return
}

func main() {

	tmpl := template.Must(template.ParseFiles("public/index.html"))

	logo, err := logo()
	if err != nil {
		fmt.Println(err)
		// fallback to 1x1 empty square
		logo = []byte("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=")
	}

	data := &Data{
		Tmpl: tmpl,
		Content: &Content{
			Logo:        logo,
			ContactMail: os.Getenv("CONTACT_MAIL"),
			Owner: &Owner{
				Name:   os.Getenv("OWNER_NAME"),
				Street: os.Getenv("OWNER_STREET"),
				Zip:    os.Getenv("OWNER_ZIP"),
				Place:  os.Getenv("OWNER_PLACE"),
			},
		},
	}

	server := &fasthttp.Server{
		Name: "provider identification",
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.SetContentType("text/html;charset=UTF-8")
			data.Content.Host = string(ctx.Host())
			data.Tmpl.Execute(ctx.Response.BodyWriter(), data.Content)
		},
	}

	if e := server.ListenAndServe("0.0.0.0:80"); e != nil {
		fmt.Printf("failed to serve: %s\n", e)
	}

}
