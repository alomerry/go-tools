package mongo

import (
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewObjectIdHex() string {
	return primitive.NewObjectID().Hex()
}

func NewObjectId() primitive.ObjectID {
	return primitive.NewObjectID()
}

func MustToObjectId(id string) primitive.ObjectID {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Panicf("invalid objectId [%v]", id)
	}
	return oid
}
