package service

type App struct {
	Team *TeamService
	User *UserService
	PR   *PRService
}

func NewApp(team *TeamService, user *UserService, pr *PRService) *App {
	return &App{
		Team: team,
		User: user,
		PR:   pr,
	}
}
