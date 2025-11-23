package service

type App struct {
	Team  *TeamService
	User  *UserService
	PR    *PRService
	Stats *StatsService
}

func NewApp(team *TeamService, user *UserService, pr *PRService, stats *StatsService) *App {
	return &App{
		Team:  team,
		User:  user,
		PR:    pr,
		Stats: stats,
	}
}
