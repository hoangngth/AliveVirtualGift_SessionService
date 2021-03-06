package auth

import (
	"AliveVirtualGift_SessionService/src/proto"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var client *redis.Client

type TokenDetails struct {
	AccessToken string
	AccessUUID  string
	AtExpires   int64
}

type AccessDetails struct {
	AccessUUID  string
	AccountID   uint64
	AccountType proto.Type
}

func InitRedis() {

	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	client = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
}

//GenerateToken ...
func GenerateToken(accInfo *proto.AccountInfo) (*TokenDetails, error) {

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 60).Unix()
	td.AccessUUID = uuid.New().String()

	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["account_id"] = accInfo.GetId()
	atClaims["account_type"] = proto.Type_value[string(accInfo.GetType())]
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	return td, nil
}

//CreateAuth ...
func CreateAuth(accountID uint64, td *TokenDetails) error {

	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	now := time.Now()

	errAccess := client.Set(td.AccessUUID, strconv.Itoa(int(accountID)), at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}

	return nil
}

//ExtractClaims ...
func ExtractClaims(tokenString string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, err
}

//ExtractTokenMetadata ...
func ExtractTokenMetadata(tokenString string) (*AccessDetails, error) {

	claims, err := ExtractClaims(tokenString)
	if err != nil {
		return nil, err
	}

	accessUUID, ok := claims["access_uuid"].(string)
	if !ok {
		return nil, err
	}

	accountID, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["account_id"]), 10, 64)
	if err != nil {
		return nil, err
	}

	accountType, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["account_id"]), 10, 32)
	if err != nil {
		return nil, err
	}

	return &AccessDetails{
		AccessUUID:  accessUUID,
		AccountID:   accountID,
		AccountType: proto.Type(accountType),
	}, nil
}

//FetchAuth ...
func FetchAuth(ad *AccessDetails) (string, error) {

	accessUUID, err := client.Get(ad.AccessUUID).Result()
	if err != nil {
		return "", err
	}

	return accessUUID, err
}

//DeleteAuth ...
func DeleteAuth(accessUUID string) (int64, error) {

	deleted, err := client.Del(accessUUID).Result()
	if err != nil {
		return 0, err
	}

	return deleted, nil
}
