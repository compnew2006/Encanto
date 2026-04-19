package api

import (
	"net/http"

	"encanto/audit"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func Router(deps Dependencies) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{deps.Config.FrontendOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		audit.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	router.Route("/api", func(api chi.Router) {
		api.Post("/auth/login", loginHandler(deps))
		api.Post("/auth/refresh", refreshHandler(deps))
		api.Post("/auth/logout", logoutHandler(deps))

		api.Group(func(protected chi.Router) {
			protected.Use(requireAuth(deps))
			protected.Use(loadCurrentUser(deps))

			protected.Get("/me", meHandler())
			protected.Get("/me/organizations", meOrganizationsHandler())
			protected.Put("/me/settings", updateSettingsHandler(deps))
			protected.Put("/me/availability", updateAvailabilityHandler(deps))
			protected.Post("/auth/switch-org", switchOrgHandler(deps))

			protected.Get("/permissions", listPermissionsHandler(deps))
			protected.Get("/roles", listRolesHandler(deps))
			protected.Post("/roles", createRoleHandler(deps))
			protected.Put("/roles/{roleID}", updateRoleHandler(deps))
			protected.Delete("/roles/{roleID}", deleteRoleHandler(deps))

			protected.Get("/users", listUsersHandler(deps))
			protected.Get("/users/{userID}", getUserHandler(deps))
			protected.Put("/users/{userID}", updateUserHandler(deps))
			protected.Get("/users/{userID}/send-restrictions", getSendRestrictionsHandler(deps))
			protected.Put("/users/{userID}/send-restrictions", updateSendRestrictionsHandler(deps))
			protected.Get("/users/{userID}/contact-visibility", getContactVisibilityHandler(deps))
			protected.Put("/users/{userID}/contact-visibility", updateContactVisibilityHandler(deps))

			protected.Get("/chats", listChatsHandler(deps))
			protected.Get("/chats/{contactID}", getChatHandler(deps))
			protected.Get("/contacts/{contactID}/messages", listMessagesHandler(deps))
			protected.Get("/contacts/{contactID}/notes", listNotesHandler(deps))
		})
	})

	return router
}
