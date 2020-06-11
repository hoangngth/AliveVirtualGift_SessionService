package auth

import (
	"AliveVirtualGift_SessionService/src/proto"
	"fmt"
	"log"
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
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUUID = uuid.New().String()

	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["account_info"] = accInfo
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

	accInfo := claims["account_info"]
	accInfoMap, ok := accInfo.(map[string]interface{})
	if !ok {
		log.Print("Payload Conversion failed")
	}

	accInfoMap["id"] = uint64(accInfoMap["id"].(float64))
	accountID := accInfoMap["id"].(uint64)

	accInfoMap["type"] = int32(accInfoMap["type"].(float64))
	accountType := proto.Type(accInfoMap["type"].(int32))

	return &AccessDetails{
		AccessUUID:  accessUUID,
		AccountID:   accountID,
		AccountType: accountType,
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
