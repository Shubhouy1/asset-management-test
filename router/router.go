package router

import (
	"github.com/Shubhouy1/asset-management/handlers"
	"github.com/Shubhouy1/asset-management/middleware"

	"github.com/go-chi/chi/v5"
)

func SetupRouter() chi.Router {

	r := chi.NewRouter()

	r.Post("/register", handlers.RegisterUser)
	r.Post("/login", handlers.LoginUser)
	r.Post("/firebase-register", handlers.FirebaseRegisterUser)

	r.Group(func(r chi.Router) {

		r.Use(middleware.AuthMiddleware)

		r.Get("/get-assets", handlers.GetTotalAssets)
		r.Post("/logout", handlers.LogoutUser)

		r.Route("/users", func(r chi.Router) {

			r.Use(middleware.RequiredRoles("admin", "asset-manager"))

			r.Get("/", handlers.GetAllUsers)
			r.Delete("/{id}", handlers.DeleteUser)

		})

		r.Route("/assets", func(r chi.Router) {

			r.Use(middleware.RequiredRoles("admin", "asset-manager"))

			r.Post("/", handlers.CreateAsset)
			r.Get("/", handlers.GetAssets)

			r.Route("/{id}", func(r chi.Router) {

				r.Put("/", handlers.UpdateAsset)
				r.Put("/assign", handlers.AssignAsset)
				r.Put("/sent-to-service", handlers.SendAssetToService)

			})

		})

	})

	return r
}
