package cmd

const (
	// Config file name and location.
	configFileName = "auth.json"
	appName        = "dbxcli"
	configBase     = ".config"

	// Environment variable names.
	tokensEnv              = "DROPBOX_TOKENS"
	personalAppKeyEnv      = "DROPBOX_PERSONAL_APP_KEY"
	personalAppSecretEnv   = "DROPBOX_PERSONAL_APP_SECRET"
	teamAccessAppKeyEnv    = "DROPBOX_TEAM_APP_KEY"
	teamAccessAppSecretEnv = "DROPBOX_TEAM_APP_SECRET"
	teamManageAppKeyEnv    = "DROPBOX_MANAGE_APP_KEY"
	teamManageAppSecretEnv = "DROPBOX_MANAGE_APP_SECRET"
)
