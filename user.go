package gosteamauth

const (
	PersonaStateOffline        int = 0
	PersonaStateOnline         int = 1
	PersonaStateBusy           int = 2
	PersonaStateAway           int = 3
	PersonaStateSnooze         int = 4
	PersonaStateLookingToTrade int = 5
	PersonaStateLookingToPlay  int = 6
)

const (
	ProfileStateConfigured    int = 1
	ProfileStateNotConfigured int = 0
)

const (
	CommunityVisibilityStatusNotVisible int = 1
	CommunityVisibilityStatusPublic     int = 3
)

// SteamUser is a steam user, as represented in the response from GetPlayerSummaries web api.
type SteamUser struct {
	// SteamID is the "steamid64" of the player.
	SteamID string `json:"steamid"`
	// PersonaName is the user's profile name.
	PersonaName string `json:"personaname"`
	// PersonaState is the current status of the user. If private, this will always be 0
	// See the PersonaState... enums.
	PersonaState int `json:"personastate"`

	// ProfileUrl is the full URL to the user's steam profile.
	ProfileUrl string `json:"profileurl"`
	// ProfileState is a int, but will always be 1 or 0 depending on if the user has their profile setup.
	// See the ProfileState... enums
	ProfileState int `json:"profilestate"`

	// CommunityVisibilityStatus represents weather the profile is visible or not, and if you're allowed to see it.
	// See the CommunityVisibilityStatus... enums
	CommunityVisibilityStatus int `json:"communityvisibilitystate"`

	// Avatar is the user's 32x32 avatar URL
	Avatar string `json:"avatar"`
	// AvatarMedium is the user's 64x64 avatar URL
	AvatarMedium string `json:"avatarmedium"`
	// AvatarFull is the user's 128x128 avatar URL
	AvatarFull string `json:"avatarfull"`
}
