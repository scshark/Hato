package app

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/dgrijalva/jwt-go"
	"github.com/scshark/Hato/internal/conf"
	"github.com/scshark/Hato/internal/model"
)

type Claims struct {
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	jwt.StandardClaims
}

func GetJWTSecret() []byte {
	return []byte(conf.JWTSetting.Secret)
}

func GenerateToken(User *model.User) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(conf.JWTSetting.Expire)
	logrus.Debugf("get expireTime %v", expireTime)

	claims := Claims{
		UID:      User.ID,
		Username: User.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    conf.JWTSetting.Issuer + ":" + User.Salt,
		},
	}
	logrus.Debugf("get claims %v", claims)

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(GetJWTSecret())
	logrus.Debugf("get token %v", token)

	return token, err
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return GetJWTSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}
