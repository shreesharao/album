package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kataras/iris/v12"
)

type RequestHandler struct {
	logger   log.Logger
	producer sarama.SyncProducer
}

func newHandler(logger log.Logger, p sarama.SyncProducer) *RequestHandler {
	return &RequestHandler{
		logger:   logger,
		producer: p,
	}
}

const (
	UploadsDir = "albums"
	KafkaTopic = "album.events"
)

//
// @Summary creates a new album
// @Description create album
// @Accept  application/x-www-form-urlencoded
// @Produce  json
// @Param  name formData string true "album name"
// @Success 200 {string} string "Album created successfully"
// @Failure 400 {string} string "Album name is required"
// @Failure 500 {string} string "Failed to create album"
// @Router /api/album/v1 [post]
func (h *RequestHandler) CreateAlbum(ctx iris.Context) {
	albumName := ctx.FormValue("name")

	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"msg": "Album name is required"})
		return
	}
	err := h.createFolder(albumName)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to create album", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Failed to create album"})
	}

	ctx.StatusCode(iris.StatusCreated)
	ctx.JSON(iris.Map{"msg": "Album created successfully", "name": albumName})
}

//
// @Summary deletes an album
// @Description delete album
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album path string true "album name"
// @Success 200 {string} string "Album deleted successfully"
// @Failure 400 {string} string "Album name is required"
// @Failure 500 {string} string "Failed to delete album"
// @Router /api/album/v1/{album} [delete]
func (h *RequestHandler) DeleteAlbum(ctx iris.Context) {
	albumName := ctx.Params().Get("album")
	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"msg": "Album name is required"})
		return
	}

	albumPath := UploadsDir + "\\" + albumName
	if _, err := os.Stat(albumPath); os.IsNotExist(err) {
		level.Error(h.logger).Log("msg", "album not found", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Album not found"})
		return
	}

	err := h.deleteFolder(albumName)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to delete album", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Failed to delete album"})
	}
	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(iris.Map{"msg": "Album deleted successfully", "name": albumName})
}

//
// @Summary creates an image
// @Description create image
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album formData string true "album name"
// @Param  file formData file true "image file"
// @Success 200 {string} string "Image uploaded successfully"
// @Failure 400 {string} string "Album name is required"
// @Failure 500 {string} string "Failed to upload image"
// @Router /api/album/image/v1/ [post]
func (h *RequestHandler) CreateImage(ctx iris.Context) {

	cntype := ctx.GetHeader("Content-Type")
	level.Info(h.logger).Log("msg", cntype)

	albumName := ctx.FormValue("album")
	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"msg": "Album name is required"})
		return
	}
	albumPath := UploadsDir + "\\" + albumName
	if _, err := os.Stat(albumPath); os.IsNotExist(err) {
		level.Error(h.logger).Log("msg", "album not found", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Album not found"})
		return
	}

	file, info, err := ctx.FormFile("file")
	if err != nil {
		level.Error(h.logger).Log("msg", "error while uploading file", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Failed to upload image"})
		return
	}

	defer file.Close()
	fname := info.Filename

	out, err := os.OpenFile(albumPath+"\\"+fname, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		level.Error(h.logger).Log("msg", "error while uploading file", "err", err.Error())

		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Failed to upload image"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		level.Error(h.logger).Log("msg", "error while uploading file", "err", err.Error())

		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Failed to upload image"})
		return
	}

	key := albumName + "-" + fname
	event := map[string]interface{}{"eventName": "ImageCreated", "album": albumName, "image": fname, "eventTime": time.Now()}
	value, err := json.Marshal(event)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to produce event to kafka", "err", err.Error())
	} else {
		h.publish(KafkaTopic, key, value)
	}

	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(iris.Map{"msg": "Image uploaded successfully", "name": fname})
}

//
// @Summary deletes an image
// @Description delete image
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album path string true "album name"
// @Param  image path string true "image name"
// @Success 200 {string} string "Image deleted successfully"
// @Failure 400 {string} string "Album name is required"
// @Failure 400 {string} string "Image name is required"
// @Failure 500 {string} string "Failed to delete image"
// @Failure 500 {string} string "Image not found"
// @Router /api/album/image/v1/{album}/{image} [delete]
func (h *RequestHandler) DeleteImage(ctx iris.Context) {
	albumName := ctx.Params().Get("album")
	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"msg": "Album name is required"})
		return
	}
	imageName := ctx.Params().Get("image")
	if imageName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"msg": "Image name is required"})
		return
	}

	imagePath := UploadsDir + "\\" + albumName + "\\" + imageName
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		level.Error(h.logger).Log("msg", "image not found", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Image not found"})
		return
	}

	err := os.Remove(imagePath)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to delete image", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Failed to delete image"})
	}

	key := albumName + "-" + imageName
	event := map[string]interface{}{"eventName": "ImageDeleted", "album": albumName, "image": imageName, "eventTime": time.Now()}
	value, err := json.Marshal(event)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to produce event to kafka", "err", err.Error())
	} else {
		h.publish(KafkaTopic, key, value)
	}

	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(iris.Map{"message": "Image deleted successfully", "name": imageName})
}

//
// @Summary returns images in an album
// @Description get images
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album path string true "album name"
// @Success 200 {object} array "[{"name":"", "url":""}]"
// @Failure 400 {string} string "Album name is required"
// @Failure 500 {string} string "Failed to get images"
// @Router /api/album/v1/{album} [get]
func (h *RequestHandler) GetImages(ctx iris.Context) {
	albumName := ctx.Params().Get("album")
	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"msg": "Album name is required"})
		return
	}

	level.Info(h.logger).Log("msg", "reading images from album", "name", albumName)
	files, err := ioutil.ReadDir(UploadsDir + "\\" + albumName)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to get images", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Failed to get images"})
	}

	level.Info(h.logger).Log("msg", "found images in album", "name", albumName, "count", len(files))

	type Image struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	images := make([]Image, 0)
	for _, f := range files {
		images = append(images, Image{Name: f.Name(), Url: "api/album/image/v1/" + albumName + "/" + f.Name()})
	}

	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(images)
}

//
// @Summary returns an image
// @Description get image
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album path string true "album name"
// @Param  image path string true "image name"
// @Success 200 {file} file "Image"
// @Failure 400 {string} string "Album name is required"
// @Failure 400 {string} string "Image name is required"
// @Failure 500 {string} string "Image not found"
// @Router /api/album/image/v1/{album}{image} [get]
func (h *RequestHandler) GetImage(ctx iris.Context) {
	albumName := ctx.Params().Get("album")
	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"msg": "Album name is required"})
		return
	}
	imageName := ctx.Params().Get("image")
	if imageName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"msg": "Image name is required"})
		return
	}

	imagePath := UploadsDir + "\\" + albumName + "\\" + imageName
	fileInfo, err := os.Stat(imagePath)
	if os.IsNotExist((err)) {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"msg": "Image not found"})
		return
	}
	level.Info(h.logger).Log("msg", "returning file", "name", imagePath)
	ctx.SendFile(imagePath, fileInfo.Name())
}

func (h *RequestHandler) ServeAPIDef(ctx iris.Context) {
	path := ".\\docs\\swagger.json"
	level.Info(h.logger).Log("msg", "serving api doc", "name", path)
	ctx.SendFile(path, "swagger.json")
}

func (h *RequestHandler) createFolder(name string) error {
	if _, err := os.Stat(UploadsDir); os.IsNotExist(err) {
		os.Mkdir(UploadsDir, 0666)
	}
	level.Info(h.logger).Log("msg", "creating album", "name", name)
	if _, err := os.Stat(UploadsDir + "\\" + name); os.IsNotExist(err) {
		err := os.Mkdir(UploadsDir+"\\"+name, 0666)
		return err
	}
	return nil
}

func (h *RequestHandler) deleteFolder(name string) error {
	level.Info(h.logger).Log("msg", "deleting album", "name", name)
	err := os.RemoveAll(UploadsDir + "\\" + name)
	return err
}

func (h *RequestHandler) publish(topic string, key string, value []byte) error {
	if h.producer == nil {
		level.Info(h.logger).Log("msg", "producer not initialized")
		return nil
	}
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	partition, offset, err := h.producer.SendMessage(message)
	if err != nil {
		return err
	}
	level.Info(h.logger).Log("msg", "publish successfull", "topic", topic, "key", key, "partition", partition, "offset", offset)
	return nil
}
