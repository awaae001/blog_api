package friendsRepositories

import "net/url"

func resolveAvatarURL(avatar string, base string) string {
	if avatar == "" {
		return ""
	}

	avatarURL, err := url.Parse(avatar)
	if err != nil {
		return avatar
	}
	if avatarURL.IsAbs() {
		return avatar
	}

	baseURL, err := url.Parse(base)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return avatar
	}

	return baseURL.ResolveReference(avatarURL).String()
}
