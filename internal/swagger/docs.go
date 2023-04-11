// Code generated by swaggo/swag. DO NOT EDIT.

package swagger

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/resources": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "tags": [
                    "ResourceAPI"
                ],
                "summary": "Query paginated resource list",
                "parameters": [
                    {
                        "type": "integer",
                        "default": 1,
                        "description": "pagination index",
                        "name": "current",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "default": 10,
                        "description": "pagination size",
                        "name": "pageSize",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "resource code (fuzzy query)",
                        "name": "code",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "resource status (enabled, disabled)",
                        "name": "status",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/utils.ResponseResult"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/schema.Resource"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "tags": [
                    "ResourceAPI"
                ],
                "summary": "Create resource record",
                "parameters": [
                    {
                        "description": "Request body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/schema.ResourceSave"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/utils.ResponseResult"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/schema.Resource"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    }
                }
            }
        },
        "/api/v1/resources/{id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "tags": [
                    "ResourceAPI"
                ],
                "summary": "Get resource details by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "unique id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/utils.ResponseResult"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/schema.Resource"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "tags": [
                    "ResourceAPI"
                ],
                "summary": "Update resource record by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "unique id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Request body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/schema.ResourceSave"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "tags": [
                    "ResourceAPI"
                ],
                "summary": "Delete resource record by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "unique id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.ResponseResult"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "errors.Error": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "detail": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "schema.Resource": {
            "type": "object",
            "properties": {
                "action": {
                    "description": "Resource action",
                    "type": "string"
                },
                "code": {
                    "description": "Unique code (format: module.resource.action)",
                    "type": "string"
                },
                "created_at": {
                    "description": "Create time",
                    "type": "string"
                },
                "description": {
                    "description": "Description",
                    "type": "string"
                },
                "id": {
                    "description": "Unique ID",
                    "type": "string"
                },
                "object": {
                    "description": "Resource object",
                    "type": "string"
                },
                "status": {
                    "description": "Status (enabled/disabled)",
                    "type": "string"
                },
                "updated_at": {
                    "description": "Update time",
                    "type": "string"
                }
            }
        },
        "schema.ResourceSave": {
            "type": "object",
            "required": [
                "action",
                "object",
                "status"
            ],
            "properties": {
                "action": {
                    "description": "Resource action",
                    "type": "string"
                },
                "code": {
                    "description": "Unique code (format: module.resource.action)",
                    "type": "string"
                },
                "description": {
                    "description": "Description",
                    "type": "string"
                },
                "object": {
                    "description": "Resource object",
                    "type": "string"
                },
                "status": {
                    "description": "Status (enabled/disabled)",
                    "type": "string",
                    "enum": [
                        "enabled",
                        "disabled"
                    ]
                }
            }
        },
        "utils.ResponseResult": {
            "type": "object",
            "properties": {
                "data": {},
                "error": {
                    "$ref": "#/definitions/errors.Error"
                },
                "success": {
                    "type": "boolean"
                },
                "total": {
                    "type": "integer"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "v10.0.0-beta",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "GIN-ADMIN",
	Description:      "RBAC scaffolding based on GIN + Gorm 2.0 + Casbin + Wire DI.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
