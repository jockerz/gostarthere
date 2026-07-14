package routes

import (
	"github.com/danielgtaylor/huma/v2"

	"vnti/apps/api/handler"
	"vnti/apps/api/schema"
)

func APICheck(v1 *huma.Group) {
	// huma.Get(v1, "/check", handler.APICheck)
	huma.Register(v1, schema.APICheck, handler.APICheck)
}

// func Home(group *huma.Group) {
// 	huma.Post(group, "/", handler.APICheck())
// }
