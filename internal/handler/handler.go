package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strconv"
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
// @Success 201 {object} Create "Album created successfully"
// @Failure 400 {object} Error "Album name is required"
// @Failure 400 {object} Error "Album is present already"
// @Failure 500 {object} Error "Failed to create album"
// @Router /api/album/v1 [post]
func (h *RequestHandler) CreateAlbum(ctx iris.Context) {
	albumName := ctx.FormValue("name")

	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		resp := Error{Msg: "Album name is required"}
		ctx.JSON(resp)
		return
	}

	albumPath := UploadsDir + "\\" + albumName
	level.Info(h.logger).Log("msg", "checking album present", "path", albumPath)
	if _, err := os.Stat(albumPath); os.IsExist(err) {
		ctx.StatusCode(iris.StatusBadRequest)
		resp := Error{Msg: "Album is present already"}
		ctx.JSON(resp)
		return
	}

	err := h.createFolder(albumName)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to create album", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		resp := Error{Msg: "Failed to create album"}
		ctx.JSON(resp)
		return
	}

	loc := "api/album/v1/" + albumName
	ctx.Header("Location", loc)
	ctx.StatusCode(iris.StatusCreated)
	resp := Create{Msg: "Album created successfully", Name: albumName, Url: loc}
	ctx.JSON(resp)
}

//
// @Summary deletes an album
// @Description delete album
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album path string true "album name"
// @Success 200 {object} Delete "Album deleted successfully"
// @Failure 400 {object} Error "Album name is required"
// @Failure 404 {object} Error "Album not found"
// @Failure 500 {object} Error "Failed to delete album"
// @Router /api/album/v1/{album} [delete]
func (h *RequestHandler) DeleteAlbum(ctx iris.Context) {
	albumName := ctx.Params().Get("album")
	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		resp := Error{Msg: "Album name is required"}
		ctx.JSON(resp)
		return
	}

	albumPath := UploadsDir + "\\" + albumName
	level.Info(h.logger).Log("msg", "checking album present", "path", albumPath)
	if _, err := os.Stat(albumPath); os.IsNotExist(err) {
		level.Error(h.logger).Log("msg", "album not found", "err", err.Error())
		ctx.StatusCode(iris.StatusNotFound)
		resp := Error{Msg: "Album not found"}
		ctx.JSON(resp)
		return
	}

	err := h.deleteFolder(albumName)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to delete album", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		resp := Error{Msg: "Failed to delete album"}
		ctx.JSON(resp)
		return
	}
	ctx.StatusCode(iris.StatusOK)
	resp := Delete{Msg: "Album deleted successfully", Name: albumName}
	ctx.JSON(resp)
}

//
// @Summary creates an image
// @Description create image
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album formData string true "album name"
// @Param  file formData file true "image file"
// @Success 201 {object} Create "Image uploaded successfully"
// @Failure 400 {object} Error "Album name is required"
// @Failure 404 {object} Error "Album not found"
// @Failure 500 {object} Error "Failed to upload image"
// @Router /api/album/image/v1/ [post]
func (h *RequestHandler) CreateImage(ctx iris.Context) {

	cntype := ctx.GetHeader("Content-Type")
	level.Info(h.logger).Log("msg", cntype)

	albumName := ctx.FormValue("album")
	if albumName == "" {

		ctx.StatusCode(iris.StatusBadRequest)
		resp := Error{Msg: "Album name is required"}
		ctx.JSON(resp)
		return
	}
	albumPath := UploadsDir + "\\" + albumName
	level.Info(h.logger).Log("msg", "checking album present", "path", albumPath)
	if _, err := os.Stat(albumPath); os.IsNotExist(err) {
		level.Error(h.logger).Log("msg", "album not found", "err", err.Error())

		resp := Error{Msg: "Album not found"}
		ctx.StatusCode(iris.StatusNotFound)
		ctx.JSON(resp)
		return
	}

	file, info, err := ctx.FormFile("file")
	if err != nil {
		level.Error(h.logger).Log("msg", "error while uploading file", "err", err.Error())

		resp := Error{Msg: "Failed to upload image"}
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(resp)
		return
	}

	defer file.Close()
	fname := info.Filename

	out, err := os.OpenFile(albumPath+"\\"+fname, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		level.Error(h.logger).Log("msg", "error while uploading file", "err", err.Error())

		resp := Error{Msg: "Failed to upload image"}
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(resp)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		level.Error(h.logger).Log("msg", "error while uploading file", "err", err.Error())

		resp := Error{Msg: "Failed to upload image"}
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(resp)
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

	loc := "api/album/image/v1/" + albumName + "/" + fname
	ctx.Header("Location", loc)
	ctx.StatusCode(iris.StatusCreated)
	resp := Create{Msg: "Image uploaded successfully", Name: fname, Url: loc}
	ctx.JSON(resp)
}

//
// @Summary deletes an image
// @Description delete image
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album path string true "album name"
// @Param  image path string true "image name"
// @Success 200 {object} Delete "Image deleted successfully"
// @Failure 400 {object} Error "Album name is required"
// @Failure 400 {object} Error "Image name is required"
// @Failure 404 {object} Error "Image not found"
// @Failure 500 {object} Error "Failed to delete image"
// @Router /api/album/image/v1/{album}/{image} [delete]
func (h *RequestHandler) DeleteImage(ctx iris.Context) {
	albumName := ctx.Params().Get("album")
	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		resp := Error{Msg: "Album name is required"}
		ctx.JSON(resp)
		return
	}
	imageName := ctx.Params().Get("image")
	if imageName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		resp := Error{Msg: "Image name is required"}
		ctx.JSON(resp)
		return
	}

	imagePath := UploadsDir + "\\" + albumName + "\\" + imageName
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		level.Error(h.logger).Log("msg", "image not found", "err", err.Error())
		ctx.StatusCode(iris.StatusNotFound)
		resp := Error{Msg: "Image not found"}
		ctx.JSON(resp)
		return
	}

	err := os.Remove(imagePath)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to delete image", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		resp := Error{Msg: "Failed to delete image"}
		ctx.JSON(resp)
		return
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
	resp := Delete{Msg: "Image deleted successfully", Name: imageName}
	ctx.JSON(resp)
}

//
// @Summary returns images in an album
// @Description get images
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album path string true "album name"
// @Success 200 {object} Get "[{"name":"", "url":""}]"
// @Failure 400 {object} Error "Album name is required"
// @Failure 404 {object} Error "Album not found"
// @Failure 500 {object} Error "Failed to get images"
// @Router /api/album/v1/{album} [get]
func (h *RequestHandler) GetImages(ctx iris.Context) {
	albumName := ctx.Params().Get("album")
	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		resp := Error{Msg: "Album name is required"}
		ctx.JSON(resp)
		return
	}

	albumPath := UploadsDir + "\\" + albumName
	level.Info(h.logger).Log("msg", "checking album present", "path", albumPath)
	if _, err := os.Stat(albumPath); os.IsNotExist(err) {
		level.Error(h.logger).Log("msg", "album not dound", "path", albumPath)

		ctx.StatusCode(iris.StatusNotFound)
		resp := Error{Msg: "Album not found"}
		ctx.JSON(resp)
		return
	}
	level.Info(h.logger).Log("msg", "reading images from album", "name", albumPath)
	files, err := ioutil.ReadDir(UploadsDir + "\\" + albumName)
	if err != nil {
		level.Error(h.logger).Log("msg", "failed to get images", "err", err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		resp := Error{Msg: "Failed to get images"}
		ctx.JSON(resp)
		return
	}

	level.Info(h.logger).Log("msg", "found images in album", "name", albumName, "count", len(files))
	if len(files) == 0 {
		ctx.StatusCode(iris.StatusOK)
		resp := Get{Msg: "No images found", Data: nil}
		ctx.JSON(resp)
		return
	}

	type Image struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	images := make([]Image, 0)
	for _, f := range files {
		images = append(images, Image{Name: f.Name(), Url: "api/album/image/v1/" + albumName + "/" + f.Name()})
	}

	ctx.StatusCode(iris.StatusOK)
	resp := Get{Msg: "Found " + strconv.Itoa(len(images)) + " image(s)", Data: images}
	ctx.JSON(resp)
}

//
// @Summary returns an image
// @Description get image
// @Accept  application/x-www-form-urlencoded
// @Produce json
// @Param  album path string true "album name"
// @Param  image path string true "image name"
// @Success 200 {file} file "Image File"
// @Failure 400 {object} Error "Album name is required"
// @Failure 400 {object} Error "Image name is required"
// @Failure 500 {object} Error "Image not found"
// @Router /api/album/image/v1/{album}{image} [get]
func (h *RequestHandler) GetImage(ctx iris.Context) {
	albumName := ctx.Params().Get("album")
	if albumName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		resp := Error{Msg: "Album name is required"}
		ctx.JSON(resp)
		return
	}
	imageName := ctx.Params().Get("image")
	if imageName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		resp := Error{Msg: "Image name is required"}
		ctx.JSON(resp)
		return
	}

	imagePath := UploadsDir + "\\" + albumName + "\\" + imageName
	fileInfo, err := os.Stat(imagePath)
	if os.IsNotExist((err)) {
		ctx.StatusCode(iris.StatusNotFound)
		resp := Error{Msg: "Image not found"}
		ctx.JSON(resp)
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
