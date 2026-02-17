package core

type BaseController struct {
	App *App
}

func NewBaseController(app *App) BaseController {
	return BaseController{App: app}
}
