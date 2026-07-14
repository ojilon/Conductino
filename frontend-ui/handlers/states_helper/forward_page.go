package states_helper

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type MakeTemplate struct {
	web_folder string
	html_file_subfolder_location string
	html_file_name string
}

func BundleSourceTemplate(webfolder string, html_file_subfolder_location string, html_file_name string) (MakeTemplate) {
	return MakeTemplate{
		web_folder: webfolder,
		html_file_subfolder_location: html_file_subfolder_location,
		html_file_name: html_file_name,
	}
}

func Push_html_file(created_template MakeTemplate, push_content any, w http.ResponseWriter) (error) {
	var pushTemplate = template.Must(template.ParseFiles(
		filepath.Join(created_template.web_folder, created_template.html_file_subfolder_location, created_template.html_file_name),
	))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := pushTemplate.Execute(w, push_content); err != nil {
		return err
	}
	return nil
}