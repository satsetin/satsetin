package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	// "github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)


func GetRegion(respw http.ResponseWriter, req *http.Request) {

	// Parse koordinat dari body request
	var longlat model.LongLat
	json.NewDecoder(req.Body).Decode(&longlat)
	// if err != nil {
	// 	var respn model.Response
	// 	respn.Status = "Error : Body tidak valid"
	// 	respn.Response = err.Error()
	// 	at.WriteJSON(respw, http.StatusBadRequest, respn)
	// 	return
	// }

	// Filter query geospasial
	filter := bson.M{
		"border": bson.M{
			"$geoIntersects": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{longlat.Longitude, longlat.Latitude},
				},
			},
		},
	}

	// Cari region berdasarkan filter
	region, err := atdb.GetOneDoc[model.Region](config.MongoconnGeo, "region", filter)
	if err != nil {
		at.WriteJSON(respw, http.StatusNotFound, bson.M{"error": "Region not found"})
		return
	}

	// Format respon sebagai FeatureCollection GeoJSON
	geoJSON := bson.M{
		"type": "FeatureCollection",
		"features": []bson.M{
			{
				"type": "Feature",
				"geometry": bson.M{
					"type":        region.Border.Type,
					"coordinates": region.Border.Coordinates,
				},
				"properties": bson.M{
					"province":     region.Province,
					"district":     region.District,
					"sub_district": region.SubDistrict,
					"village":      region.Village,
				},
			},
		},
	}

	// Kirim respon dalam format GeoJSON
	at.WriteJSON(respw, http.StatusOK, geoJSON)
}

//GET ROADS
func GetRoads(respw http.ResponseWriter, req *http.Request) {
	var longlat model.LongLat
	// Pastikan body request didekode dengan benar
	if err := json.NewDecoder(req.Body).Decode(&longlat); err != nil {
		at.WriteJSON(respw, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Filter untuk query MongoDB berdasarkan lokasi dan jarak
	filter := bson.M{
		"geometry": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",  // Menambahkan tipe "Point" untuk GeoJSON
					"coordinates": []float64{longlat.Longitude, longlat.Latitude}, // Koordinat
				},
				"$maxDistance": longlat.MaxDistance, // Maksimum jarak
			},
		},
	}

	// Ambil data jalan dari MongoDB
	roads, err := atdb.GetAllDoc[[]model.Roads](config.MongoconnGeo, "roads", filter)
	if err != nil {
		at.WriteJSON(respw, http.StatusNotFound, "Data not found")
		return
	}

	// Kirim respons dengan format GeoJSON yang benar
	// Pastikan objek roads memiliki struktur yang sesuai dengan format GeoJSON
	geoJSONResponse := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": make([]interface{}, len(roads)),
	}

	for i, road := range roads {
		geoJSONResponse["features"].([]interface{})[i] = map[string]interface{}{
			"type": "Feature",
			"geometry": map[string]interface{}{
				"type":        "LineString", // Sesuaikan tipe geometry dengan data yang sesuai, misalnya "LineString" untuk jalan
				"coordinates": road.Geometry.Coordinates, // Pastikan data koordinat jalan tersedia dalam format array
			},
			"properties": road.Properties, // Misalkan ada properties terkait jalan yang perlu dikirim
		}
	}

	// Kirimkan respons GeoJSON
	at.WriteJSON(respw, http.StatusOK, geoJSONResponse)
}
