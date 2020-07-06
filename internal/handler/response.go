package handler

type Create struct {
	Msg  string `json:"msg"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Delete struct {
	Msg  string `json:"msg"`
	Name string `json:"name"`
}

type Error struct {
	Msg string `json:"msg"`
}

type Get struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
