package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"vnti/pkg/entities"
	"vnti/pkg/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const CtxUserKey contextKey = "user"
const BearerPrefix = "Bearer "

func Auth(jwtSecret string, authSrv service.AuthService, userSrv service.UserService) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		jwtTokenStr, err := getAuthToken(ctx.Header("Authorization"))
		if err != nil {
			writeUnauthorized(ctx, err.Error())
			return
		}

		claims := entities.JWTClaims{}
		token, err := jwt.ParseWithClaims(jwtTokenStr, &claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			// _, ok := token.Claims.(entities.JWTClaims)
			// if !ok {
			// 	return nil, errors.New("unexpected token claims")
			// }
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			writeUnauthorized(ctx, "invalid or expired token")
			return
		}

		tokenStr := token.Claims.(*entities.JWTClaims).ID
		user, err := authSrv.AuthByToken(ctx.Context(), tokenStr)
		if err != nil {
			writeUnauthorized(ctx, "invalid user")
			return
		} else if !user.Active {
			writeUnauthorized(ctx, "inactive user. activation is required")
			return
		}

		ctx = huma.WithValue(ctx, CtxUserKey, user)
		next(ctx)
	}
}

func AuthAllowNonActive(jwtSecret string, authSrv service.AuthService, userSrv service.UserService) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		jwtTokenStr, err := getAuthToken(ctx.Header("Authorization"))
		if err != nil {
			writeUnauthorized(ctx, err.Error())
			return
		}

		claims := entities.JWTClaims{}
		token, err := jwt.ParseWithClaims(jwtTokenStr, &claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			// _, ok := token.Claims.(entities.JWTClaims)
			// if !ok {
			// 	return nil, errors.New("unexpected token claims")
			// }
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			writeUnauthorized(ctx, "invalid or expired token")
			return
		}

		tokenStr := token.Claims.(*entities.JWTClaims).ID
		user, err := authSrv.AuthByToken(ctx.Context(), tokenStr)
		if err != nil {
			writeUnauthorized(ctx, "invalid user")
			return
		}

		ctx = huma.WithValue(ctx, CtxUserKey, user)
		next(ctx)
	}
}

// Get token from Authorization header
func getAuthToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}
	if len(authHeader) < len(BearerPrefix) || authHeader[:len(BearerPrefix)] != BearerPrefix {
		return "", errors.New("invalid authorization header format")
	}
	return authHeader[len(BearerPrefix):], nil
}

func writeUnauthorized(ctx huma.Context, msg string) {
	ctx.SetStatus(http.StatusUnauthorized)
	ctx.SetHeader("Content-Type", "application/json")
	body, _ := json.Marshal(map[string]any{
		"success": false,
		"message": msg,
		"data":    nil,
	})
	_, _ = ctx.BodyWriter().Write(body)
}
