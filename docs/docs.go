// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/album/image/v1/": {
            "post": {
                "description": "create image",
                "consumes": [
                    "application/x-www-form-urlencoded"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "creates an image",
                "parameters": [
                    {
                        "type": "string",
                        "description": "album name",
                        "name": "album",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "file",
                        "description": "image file",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Image uploaded successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Album name is required",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to upload image",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/album/image/v1/{album}/{image}": {
            "delete": {
                "description": "delete image",
                "consumes": [
                    "application/x-www-form-urlencoded"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "deletes an image",
                "parameters": [
                    {
                        "type": "string",
                        "description": "album name",
                        "name": "album",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "image name",
                        "name": "image",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Image deleted successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Image name is required",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Image not found",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/album/image/v1/{album}{image}": {
            "get": {
                "description": "get image",
                "consumes": [
                    "application/x-www-form-urlencoded"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "returns an image",
                "parameters": [
                    {
                        "type": "string",
                        "description": "album name",
                        "name": "album",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "image name",
                        "name": "image",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Image",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "400": {
                        "description": "Image name is required",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Image not found",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/album/v1": {
            "post": {
                "description": "create album",
                "consumes": [
                    "application/x-www-form-urlencoded"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "creates a new album",
                "parameters": [
                    {
                        "type": "string",
                        "description": "album name",
                        "name": "name",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Album created successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Album name is required",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to create album",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/album/v1/{album}": {
            "get": {
                "description": "get images",
                "consumes": [
                    "application/x-www-form-urlencoded"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "returns images in an album",
                "parameters": [
                    {
                        "type": "string",
                        "description": "album name",
                        "name": "album",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "[{\"name\":\"\", \"url\":\"\"}]",
                        "schema": {
                            "type": "array"
                        }
                    },
                    "400": {
                        "description": "Album name is required",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to get images",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "description": "delete album",
                "consumes": [
                    "application/x-www-form-urlencoded"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "deletes an album",
                "parameters": [
                    {
                        "type": "string",
                        "description": "album name",
                        "name": "album",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Album deleted successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Album name is required",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to delete album",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}