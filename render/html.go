package render

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
)

type HTMLProduction struct {
	Template *template.Template
	Name     string
	FuncMap  template.FuncMap
	Data     any
}

type HTMLDevelopment struct {
	template     *template.Template
	TemplatePath string
	TemplateName string
	FuncMap      template.FuncMap
	Data         any
}

var htmlContentType = []string{"text/html; charset=utf-8"}

func (h HTMLProduction) Render(w http.ResponseWriter) error {
	if h.Template == nil {
		tmpl := template.New(filepath.Base(h.Name))
		if h.FuncMap != nil {
			tmpl = tmpl.Funcs(h.FuncMap)
		}

		parsedTemplate, err := tmpl.ParseFiles(h.Name)
		if err != nil {
			http.Error(w, "template parse error: "+err.Error(), http.StatusInternalServerError)
			return err
		}
		h.Template = parsedTemplate
	}

	h.WritingContentType(w)

	if h.Name != "" {
		return h.Template.ExecuteTemplate(w, h.Name, h.Data)
	}

	return h.Template.Execute(w, h.Data)
}

func (h HTMLProduction) WritingContentType(w http.ResponseWriter) error {
	writeContentType(w, htmlContentType)
	return nil
}

func (hd *HTMLDevelopment) Render(w http.ResponseWriter) error {
	if hd.TemplatePath == "" {
		http.Error(w, "template path is empty", http.StatusInternalServerError)
		return errors.New("template path is empty")
	}

	hd.WritingContentType(w)

	if hd.template == nil {
		tmpl := template.New(filepath.Base(hd.TemplatePath))
		if hd.FuncMap != nil {
			tmpl = tmpl.Funcs(hd.FuncMap)
		}

		parsedTemplate, err := tmpl.ParseFiles(hd.TemplatePath)
		if err != nil {
			http.Error(w, "template parse error: "+err.Error(), http.StatusInternalServerError)
			return err
		}

		hd.template = parsedTemplate
	}

	if hd.TemplateName != "" {
		return hd.template.ExecuteTemplate(w, hd.TemplateName, hd.Data)
	}

	return hd.template.Execute(w, hd.Data)
}

func (hd HTMLDevelopment) WritingContentType(w http.ResponseWriter) error {
	writeContentType(w, htmlContentType)
	return nil
}
