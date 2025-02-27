package main

import (
	// postgreSQL integration

	"database/sql"
	"encoding/json"
	"github.com/beeker1121/goque"

	"go.uber.org/zap"
)

type stateQueue struct {
	DBAssetID   int
	State       int
	TimestampMs int64
}
type state struct {
	State       int   `json:"state"`
	TimestampMs int64 `json:"timestamp_ms"`
}

// ProcessStateData processes an incoming state message
func ProcessStateData(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload state

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)

	newObject := stateQueue{
		TimestampMs: parsedPayload.TimestampMs,
		State:       parsedPayload.State,
		DBAssetID:   DBassetID,
	}

	_, err = pg.EnqueueObject([]byte(prefixState), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}

}

type countQueue struct {
	DBAssetID   int
	Count       int
	Scrap       int
	TimestampMs int64
}
type count struct {
	Count       int   `json:"count"`
	Scrap       int   `json:"scrap"`
	TimestampMs int64 `json:"timestamp_ms"`
}

// ProcessCountData processes an incoming count message
func ProcessCountData(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload count

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	// this should not happen. Throw a warning message and ignore (= do not try to store in database)
	if parsedPayload.Count <= 0 {
		zap.S().Warnf("count <= 0", customerID, location, assetID, payload, parsedPayload)
		return
	}

	DBassetID := GetAssetID(customerID, location, assetID)

	newObject := countQueue{
		TimestampMs: parsedPayload.TimestampMs,
		Count:       parsedPayload.Count,
		Scrap:       parsedPayload.Scrap,
		DBAssetID:   DBassetID,
	}

	_, err = pg.EnqueueObject([]byte(prefixCount), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}

}

type scrapCountQueue struct {
	DBAssetID   int
	Scrap       int
	TimestampMs int64
}
type scrapCount struct {
	Scrap       int   `json:"scrap"`
	TimestampMs int64 `json:"timestamp_ms"`
}

// ProcessScrapCountData processes an incoming scrapCount message
func ProcessScrapCountData(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload scrapCount

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)

	newObject := scrapCountQueue{
		TimestampMs: parsedPayload.TimestampMs,
		Scrap:       parsedPayload.Scrap,
		DBAssetID:   DBassetID,
	}

	_, err = pg.EnqueueObject([]byte(prefixScrapCount), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}
}

type addShiftQueue struct {
	DBAssetID      int
	TimestampMs    int64
	TimestampMsEnd int64
}
type addShift struct {
	TimestampMs    int64 `json:"timestamp_ms"`
	TimestampMsEnd int64 `json:"timestamp_ms_end"`
}

// ProcessAddShift adds a new shift to the database
func ProcessAddShift(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload addShift

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)
	newObject := addShiftQueue{
		TimestampMs:    parsedPayload.TimestampMs,
		TimestampMsEnd: parsedPayload.TimestampMsEnd,
		DBAssetID:      DBassetID,
	}

	_, err = pg.EnqueueObject([]byte(prefixAddShift), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}
}

type addMaintenanceActivityQueue struct {
	DBAssetID     int
	TimestampMs   int64
	ComponentName string
	Activity      int
	ComponentID   int
}
type addMaintenanceActivity struct {
	TimestampMs   int64  `json:"timestamp_ms"`
	ComponentName string `json:"component"`
	Activity      int    `json:"activity"`
}

// ProcessAddMaintenanceActivity adds a new maintenance activity to the database
func ProcessAddMaintenanceActivity(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload addMaintenanceActivity

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)

	componentID := GetComponentID(DBassetID, parsedPayload.ComponentName)
	if componentID == 0 {
		zap.S().Errorf("GetComponentID failed")
		return
	}

	newObject := addMaintenanceActivityQueue{
		DBAssetID:     DBassetID,
		TimestampMs:   parsedPayload.TimestampMs,
		ComponentName: parsedPayload.ComponentName,
		ComponentID:   componentID,
		Activity:      parsedPayload.Activity,
	}
	_, err = pg.EnqueueObject([]byte(prefixAddMaintenanceActivity), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}

}

type uniqueProductQueue struct {
	DBAssetID        int
	UID              string
	TimestampMsBegin int64
	TimestampMsEnd   int64
	ProductID        string
	IsScrap          bool
	QualityClass     string
	StationID        string
}
type uniqueProduct struct {
	UID              string `json:"UID"`
	TimestampMsBegin int64  `json:"begin_timestamp_ms"`
	TimestampMsEnd   int64  `json:"end_timestamp_ms"`
	ProductID        string `json:"productID"`
	IsScrap          bool   `json:"isScrap"`
	QualityClass     string `json:"qualityClass"`
	StationID        string `json:"stationID"`
}

// ProcessUniqueProduct adds a new uniqueProduct to the database
func ProcessUniqueProduct(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload uniqueProduct

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)
	newObject := uniqueProductQueue{
		DBAssetID:        DBassetID,
		UID:              parsedPayload.UID,
		TimestampMsBegin: parsedPayload.TimestampMsBegin,
		TimestampMsEnd:   parsedPayload.TimestampMsEnd,
		ProductID:        parsedPayload.ProductID,
		IsScrap:          parsedPayload.IsScrap,
		QualityClass:     parsedPayload.QualityClass,
		StationID:        parsedPayload.StationID,
	}

	_, err = pg.EnqueueObject([]byte(prefixUniqueProduct), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}
}

type scrapUniqueProductQueue struct {
	DBAssetID int
	UID       string
}
type scrapUniqueProduct struct {
	UID string `json:"UID"`
}

// ProcessScrapUniqueProduct sets isScrap of a uniqueProduct to true
func ProcessScrapUniqueProduct(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload scrapUniqueProduct

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)
	newObject := scrapUniqueProductQueue{
		UID:       parsedPayload.UID,
		DBAssetID: DBassetID,
	}

	_, err = pg.EnqueueObject([]byte(prefixUniqueProductScrap), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}
}

type addProductQueue struct {
	DBAssetID            int
	ProductName          string
	TimePerUnitInSeconds float64
}
type addProduct struct {
	ProductName          string  `json:"product_id"`
	TimePerUnitInSeconds float64 `json:"time_per_unit_in_seconds"`
}

// ProcessAddProduct adds a new product to the database
func ProcessAddProduct(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload addProduct

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)
	newObject := addProductQueue{
		DBAssetID:            DBassetID,
		ProductName:          parsedPayload.ProductName,
		TimePerUnitInSeconds: parsedPayload.TimePerUnitInSeconds,
	}

	_, err = pg.EnqueueObject([]byte(prefixAddProduct), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}
}

type addOrderQueue struct {
	DBAssetID   int
	ProductName string
	OrderName   string
	TargetUnits int
	ProductID   int
}
type addOrder struct {
	ProductName string `json:"product_id"`
	OrderName   string `json:"order_id"`
	TargetUnits int    `json:"target_units"`
}

// ProcessAddOrder adds a new order without begin and end timestamp to the database
func ProcessAddOrder(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload addOrder

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)

	productID, err := GetProductID(DBassetID, parsedPayload.ProductName)
	if err == sql.ErrNoRows {
		zap.S().Errorf("Product does not exist yet", DBassetID, parsedPayload.ProductName, parsedPayload.OrderName)
		return
	} else if err != nil { // never executed
		PQErrorHandling("GetProductID db.QueryRow()", err)
	}

	newObject := addOrderQueue{
		DBAssetID:   DBassetID,
		ProductName: parsedPayload.ProductName,
		OrderName:   parsedPayload.OrderName,
		TargetUnits: parsedPayload.TargetUnits,
		ProductID:   productID,
	}
	_, err = pg.EnqueueObject([]byte(prefixAddOrder), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}

}

type startOrderQueue struct {
	DBAssetID   int
	TimestampMs int64
	OrderName   string
}
type startOrder struct {
	TimestampMs int64  `json:"timestamp_ms"`
	OrderName   string `json:"order_id"`
}

// ProcessStartOrder starts an order by adding beginTimestamp
func ProcessStartOrder(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload startOrder

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)
	newObject := startOrderQueue{
		TimestampMs: parsedPayload.TimestampMs,
		OrderName:   parsedPayload.OrderName,
		DBAssetID:   DBassetID,
	}

	_, err = pg.EnqueueObject([]byte(prefixStartOrder), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}
}

type endOrderQueue struct {
	DBAssetID   int
	TimestampMs int64
	OrderName   string
}
type endOrder struct {
	TimestampMs int64  `json:"timestamp_ms"`
	OrderName   string `json:"order_id"`
}

// ProcessEndOrder starts an order by adding endTimestamp
func ProcessEndOrder(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload endOrder

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)
	newObject := endOrderQueue{
		TimestampMs: parsedPayload.TimestampMs,
		OrderName:   parsedPayload.OrderName,
		DBAssetID:   DBassetID,
	}

	_, err = pg.EnqueueObject([]byte(prefixEndOrder), newObject)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}
}

type recommendationStruct struct {
	UID                  string
	TimestampMs          int64 `json:"timestamp_ms"`
	Customer             string
	Location             string
	Asset                string
	RecommendationType   int
	Enabled              bool
	RecommendationValues string
	DiagnoseTextDE       string
	DiagnoseTextEN       string
	RecommendationTextDE string
	RecommendationTextEN string
}

// ProcessRecommendationData processes an incoming count message
func ProcessRecommendationData(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload recommendationStruct

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	_, err = pg.EnqueueObject([]byte(prefixRecommendation), parsedPayload)
	if err != nil {
		zap.S().Errorf("Error enqueuing", err)
		return
	}
}

type processValueQueue struct {
	DBAssetID   int
	TimestampMs int64
	Name        string
	Value       int
}

type processValueFloat64Queue struct {
	DBAssetID   int
	TimestampMs int64
	Name        string
	Value       float64
}

// ProcessProcessValueData processes an incoming processValue message
func ProcessProcessValueData(customerID string, location string, assetID string, payloadType string, payload []byte, pg *goque.PrefixQueue) {
	var parsedPayload interface{}

	err := json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		zap.S().Errorf("json.Unmarshal failed", err, payload)
	}

	DBassetID := GetAssetID(customerID, location, assetID)

	// process unknown data structure according to https://blog.golang.org/json
	m := parsedPayload.(map[string]interface{})

	if val, ok := m["timestamp_ms"]; ok { //if timestamp_ms key exists (https://stackoverflow.com/questions/2050391/how-to-check-if-a-map-contains-a-key-in-go)
		timestampMs, ok := val.(int64)
		if !ok {
			timestampMsFloat, ok2 := val.(float64)
			if !ok2 {
				zap.S().Errorf("Timestamp not int64 nor float64", payload, val)
				return
			}
			timestampMs = int64(timestampMsFloat)
		}

		// loop through map
		for k, v := range m {
			switch k {
			case "timestamp_ms":
			case "measurement":
			case "serial_number":
				break //ignore them
			default:
				value, ok := v.(int)
				if !ok {
					valueFloat64, ok2 := v.(float64)
					if !ok2 {
						zap.S().Errorf("Process value recieved that is not an integer nor float", k, v)
						break
					}
					newObject := processValueFloat64Queue{
						DBAssetID:   DBassetID,
						TimestampMs: timestampMs,
						Name:        k,
						Value:       valueFloat64,
					}
					_, err = pg.EnqueueObject([]byte(prefixProcessValueFloat64), newObject)
					if err != nil {
						zap.S().Errorf("Error enqueuing", err)
						return
					}
					break
				}
				newObject := processValueQueue{
					DBAssetID:   DBassetID,
					TimestampMs: timestampMs,
					Name:        k,
					Value:       value,
				}
				_, err = pg.EnqueueObject([]byte("processValue"), newObject)
				if err != nil {
					zap.S().Errorf("Error enqueuing", err)
					return
				}
			}
		}

	}
}
