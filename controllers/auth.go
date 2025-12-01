package controllers

import (
	"encoding/json"
	"fmt"
	"hyperlane/logger"
	"hyperlane/models"
	"hyperlane/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func HandleLogin(c *gin.Context) {
	var req SignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Errorf("Invalid request: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request. Please try again later.", nil)
		return
	}

	loginResp, err := processOAuthLogin(req.Code)
	if err != nil {
		logger.Log.Errorf("Login failed: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "success", loginResp)
}

// HandleOAuthCallback 处理 OAuth GET 回调请求
func HandleOAuthCallback(c *gin.Context) {
	frontendUrl := viper.GetString("app.frontendUrl")
	if frontendUrl == "" {
		frontendUrl = "https://www.hyperlane.cc"
	}

	// 从 URL 查询参数获取 code
	code := c.Query("code")
	if code == "" {
		logger.Log.Error("Missing code parameter in OAuth callback")
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=missing_code", frontendUrl))
		return
	}

	loginResp, err := processOAuthLogin(code)
	if err != nil {
		logger.Log.Errorf("OAuth login failed: %v", err)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=login_failed", frontendUrl))
		return
	}

	// 成功：重定向到首页并带上 token
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/?token=%s", frontendUrl, loginResp.Token))
}

// processOAuthLogin 封装通用的 OAuth 登录逻辑
func processOAuthLogin(code string) (*LoginResponse, error) {
	var accessRequest AccessTokenRequest
	accessRequest.ClientId = viper.GetString("oauth.clientId")
	accessRequest.ClientSecret = viper.GetString("oauth.clientSecret")
	accessRequest.Code = code
	accessRequest.RedirectUri = viper.GetString("oauth.redirectUri")

	var reqArgs utils.HTTPRequestParams
	reqArgs.URL = viper.GetString("oauth.accessApi")
	reqArgs.Method = "POST"
	reqArgs.Body = accessRequest

	// GitHub OAuth requires Accept header for JSON response
	header := make(map[string]string)
	header["Accept"] = "application/json"
	reqArgs.Headers = header

	result, err := utils.SendHTTPRequest(reqArgs)
	if err != nil {
		return nil, fmt.Errorf("network error")
	}

	// Parse GitHub OAuth response
	var tokenResp GitHubAccessTokenResponse
	err = json.Unmarshal([]byte(result), &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse oauth response")
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("invalid oauth response")
	}

	accessToken := tokenResp.AccessToken
	reqArgs.URL = viper.GetString("oauth.getUser")
	reqArgs.Method = "GET"

	// Reuse header map for user API request
	header = make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", accessToken)
	reqArgs.Headers = header

	userResult, err := utils.SendHTTPRequest(reqArgs)
	if err != nil {
		return nil, fmt.Errorf("network error")
	}

	// Parse GitHub User response
	var githubUser GitHubUserResponse
	err = json.Unmarshal([]byte(userResult), &githubUser)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user data")
	}

	if githubUser.ID == 0 {
		return nil, fmt.Errorf("invalid user data")
	}

	var user *models.User

	// Use GitHub ID as Uid
	userId := uint(githubUser.ID)
	user, err = models.GetUserByUid(userId)
	if err == nil {
		// Update existing user
		user.Uid = userId
		user.Email = githubUser.Email
		user.Github = githubUser.HTMLURL
		err = models.UpdateUser(user)
	} else {
		// Create new user
		var u models.User
		u.Uid = userId
		u.Avatar = githubUser.AvatarURL
		u.Email = githubUser.Email
		u.Username = githubUser.Login
		if githubUser.Name != "" {
			u.Username = githubUser.Name
		}
		u.Github = githubUser.HTMLURL
		user = &u
		err = models.CreateUser(user)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save user")
	}

	// TODO: gocache?
	perms, err := models.GetUserWithPermissions(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Avatar, user.Username, user.Github, perms)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token")
	}

	return &LoginResponse{
		User:        *user,
		Permissions: perms,
		Token:       token,
	}, nil
}
