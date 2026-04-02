package mongo

import (
  "log"
  
  "go.mongodb.org/mongo-driver/v2/bson"
)

func NewObjectIdHex() string {
  return bson.NewObjectID().Hex()
}

func NewObjectId() bson.ObjectID {
  return bson.NewObjectID()
}

func MustToObjectId(id string) bson.ObjectID {
  oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		log.Panicf("invalid objectId [%v]", id)
	}
	return oid
}
