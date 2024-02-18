package unit_test

import (
	"context"
	"encoding/json"
	"fmt"
	"main/cache"
	"main/utils"
	"math/rand"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
)

func TestSetCache(t *testing.T) {
	rc, mock := redismock.NewClientMock()
	cache.RedisClient = rc

	newKey := "http://localhost:8080/api/v1/ad?age=20&gender=M&country=US"
	newValue := []utils.Item{
		{Title: "Test1", EndAt: time.Now().Add(time.Duration(rand.Intn(86400)) * time.Second)},
		{Title: "Test2", EndAt: time.Now().Add(time.Duration(rand.Intn(86400)) * time.Second)},
		{Title: "Test3", EndAt: time.Now().Add(time.Duration(rand.Intn(86400)) * time.Second)},
	}
	jsonValue, _ := json.Marshal(newValue)

	// Normal case
	mock.ExpectSet(newKey, string(jsonValue), 5*time.Minute).SetVal("OK")

	err := cache.SetCache(context.Background(), newKey, newValue)

	if err != nil {
		fmt.Println(err)
		t.Errorf("Error was not expected while setting cache")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	// Error case
	mock.ExpectSet(newKey, string(jsonValue), 5*time.Minute).SetErr(fmt.Errorf("error setting cache"))

	err = cache.SetCache(context.Background(), newKey, newValue)

	if err.Error() != "error setting cache" {
		t.Errorf("Error was expected while setting cache")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestAddConditionCache(t *testing.T) {
	rc, mock := redismock.NewClientMock()
	cache.RedisClient = rc

	conditionKind := "age"
	newKey := "http://localhost:8080/api/v1/ad?age=20&gender=M&country=US"

	// Normal case
	mock.ExpectLPush(conditionKind, newKey).SetVal(1)

	err := cache.AddConditionCache(context.Background(), conditionKind, newKey)

	if err != nil {
		t.Errorf("Error was not expected while adding condition cache")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	// Error case
	mock.ExpectLPush(conditionKind, newKey).SetErr(fmt.Errorf("error adding condition cache"))

	err = cache.AddConditionCache(context.Background(), conditionKind, newKey)

	if err.Error() != "error adding condition cache" {
		t.Errorf("Error was expected while adding condition cache")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestDeleteConditionCache(t *testing.T) {
	rc, mock := redismock.NewClientMock()
	cache.RedisClient = rc

	mockKeys := []string{
		"http://localhost:8080/api/v1/ad?age=20",
		"http://localhost:8080/api/v1/ad?age=30",
		"http://localhost:8080/api/v1/ad?age=23",
	}

	// Normal case
	mock.ExpectLRange("age", 0, -1).SetVal(mockKeys)
	for _, k := range mockKeys {
		mock.ExpectDel(k).SetVal(1)
	}
	mock.ExpectDel("age").SetVal(1)

	err := cache.DeleteConditionCache(context.Background(), "age")

	if err != nil {
		t.Errorf("Error was not expected while deleting condition cache")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	// Error case 1
	mock.ExpectLRange("age", 0, -1).SetErr(fmt.Errorf("error fetching kind from redis"))
	err = cache.DeleteConditionCache(context.Background(), "age")

	if err.Error() != "error fetching kind from redis" {
		t.Errorf("Error was expected while deleting condition cache")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	// Error case 2
	mock.ExpectLRange("age", 0, -1).SetVal(mockKeys)
	mock.ExpectDel(mockKeys[0]).SetErr(fmt.Errorf("error fetching data from redis"))

	err = cache.DeleteConditionCache(context.Background(), "age")

	if err.Error() != "error fetching data from redis" {
		t.Errorf("Error was expected while deleting condition cache")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
