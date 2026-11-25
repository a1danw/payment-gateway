package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	pb "github.com/a1danw/gomicro/proto"
	jwt "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// access db setup in auth/cmd/auth/server.go - db, err = sql.Open(dbDriver, dsn)
// from oop perspective this is the class which implements the AuthService from auth_svc.proto (which generates the interface in auth_svc_grpc.pb.go)
// db passed into the constructor Implementation
type Implementation struct {
	db *sql.DB
	pb.UnimplementedAuthServiceServer
}

// create instance of Implementation - esentially constructor for the Implementation struct - pointer to the Implementation
// in auth/cmd/auth/server.go - allows us to create a db instance - authServerImplementation := auth.NewAuthImplementation(db)
func NewAuthImplementation(db *sql.DB) *Implementation {
	return &Implementation{db: db}
}

// implementing the interface from auth_svg.proto template which generated auth_svc_grpc.pb.go - AuthServiceClient
// Credentials - comes from auth_svc.pb.go which generated the Credentials, Token and User structs
func (this *Implementation) GetToken(ctx context.Context, credentials *pb.Credentials) (*pb.Token, error) {
	type user struct {
		userId   string
		password string
	}

	var u user
	stmt, err := this.db.Prepare("SELECT user_id, password from user WHERE user_id=? AND password=?")
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// query db and scan into user struct - fill in ? placeholders
	err = stmt.QueryRow(credentials.GetUserName(), credentials.GetPassword()).Scan(&u.userId, &u.password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		// internal error
		return nil, status.Error(codes.Internal, err.Error())
	}
	// return jwt for authenticated user
	jwToken, err := createJWT(u.userId)
	if err != nil {
		return nil, err
	}

	// return token to client
	return &pb.Token{Jwt: jwToken}, nil
}

func createJWT(userID string) (string, error) {
	// single private key
	key := []byte(os.Getenv("SIGNING_KEY"))
	now := time.Now()

	// signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": "auth-service",
			"sub": userID,
			"iat": now.Unix(),
			"exp": now.Add(24 * time.Hour).Unix(),
		})

	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	return signedToken, nil
}

func (this *Implementation) ValidateToken(ctx context.Context, token *pb.Token) (*pb.User, error) {
	key := []byte(os.Getenv("SIGNING_KEY"))

	userId, err := validateJWT(token.Jwt, key)
	if err != nil {
		return nil, err
	}
	// use the user id in the claims of the token
	return &pb.User{UserId: userId}, nil
}

func validateJWT(t string, signingKey []byte) (string, error) {
	type MyClaims struct {
		jwt.RegisteredClaims
	}

	// parse the token
	parsedToken, err := jwt.ParseWithClaims(t, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		// if token is expired
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", status.Error(codes.Unauthenticated, "token expired")
			// anything else
		} else {
			return "", status.Error(codes.Unauthenticated, "unauthenticated")
		}
	}
	// type assertion of MyClaims - get claim out of parsed token
	claims, ok := parsedToken.Claims.(*MyClaims)
	if !ok {
		return "", status.Error(codes.Internal, "claims type assertion failed")
	}

	// return user id
	return claims.RegisteredClaims.Subject, nil
}
