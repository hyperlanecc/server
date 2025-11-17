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

	var accessRequest AccessTokenRequest
	accessRequest.ClientId = viper.GetString("oauth.clientId")
	accessRequest.ClientSecret = viper.GetString("oauth.clientSecret")
	accessRequest.Code = req.Code

	var reqArgs utils.HTTPRequestParams
	reqArgs.URL = viper.GetString("oauth.accessApi")
	reqArgs.Method = "POST"
	reqArgs.Body = accessRequest

	// GitHub OAuth requires Accept header for JSON response
	header := make(map[string]string)
	header["Accept"] = "application/json"
	reqArgs.Headers = header

	var result string
	var err error
	result, err = utils.SendHTTPRequest(reqArgs)
	if err != nil {
		logger.Log.Errorf("ServerError: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Network error, please try again later.", nil)
		return
	}

	// Parse GitHub OAuth response
	var tokenResp GitHubAccessTokenResponse
	err = json.Unmarshal([]byte(result), &tokenResp)
	if err != nil {
		logger.Log.Errorf("Failed to parse access token response: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to parse OAuth response", nil)
		return
	}

	if tokenResp.AccessToken == "" {
		logger.Log.Errorf("Empty access token received: %v", tokenResp)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid OAuth response", nil)
		return
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
		logger.Log.Errorf("SendHTTPRequest err: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Network error, please try again later.", nil)
		return
	}

	// Parse GitHub User response
	var githubUser GitHubUserResponse
	err = json.Unmarshal([]byte(userResult), &githubUser)
	if err != nil {
		logger.Log.Errorf("Failed to parse user response: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to parse user data", nil)
		return
	}

	if githubUser.ID == 0 {
		logger.Log.Errorf("Invalid user data received: %v", githubUser)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user data", nil)
		return
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
		logger.Log.Errorf("ServerError: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Network error, please try again later.", nil)
		return
	}

	// TODO: gocache?
	perms, err := models.GetUserWithPermissions(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "get permissions error", nil)
		return
	}

	var loginResp LoginResponse
	loginResp.User = *user
	loginResp.Permissions = perms

	token, err := utils.GenerateToken(user.ID, user.Email, user.Avatar, user.Username, user.Github, perms)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "generate token error", nil)
		return
	}
	loginResp.Token = token
	utils.SuccessResponse(c, http.StatusOK, "success", loginResp)
}

// HandleOAuthCallback 处理 OAuth GET 回调请求
func HandleOAuthCallback(c *gin.Context) {
	// 从 URL 查询参数获取 code
	code := c.Query("code")
	if code == "" {
		logger.Log.Error("Missing code parameter in OAuth callback")
		c.Redirect(http.StatusFound, "https://www.hyperlane.cc/login?error=missing_code")
		return
	}

	// 构造登录请求
	var accessRequest AccessTokenRequest
	accessRequest.ClientId = viper.GetString("oauth.clientId")
	accessRequest.ClientSecret = viper.GetString("oauth.clientSecret")
	accessRequest.Code = code

	var reqArgs utils.HTTPRequestParams
	reqArgs.URL = viper.GetString("oauth.accessApi")
	reqArgs.Method = "POST"
	reqArgs.Body = accessRequest

	header := make(map[string]string)
	header["Accept"] = "application/json"
	reqArgs.Headers = header

	result, err := utils.SendHTTPRequest(reqArgs)
	if err != nil {
		logger.Log.Errorf("OAuth access token error: %v", err)
		c.Redirect(http.StatusFound, "https://www.hyperlane.cc/login?error=oauth_failed")
		return
	}

	var tokenResp GitHubAccessTokenResponse
	err = json.Unmarshal([]byte(result), &tokenResp)
	if err != nil || tokenResp.AccessToken == "" {
		logger.Log.Errorf("Invalid OAuth token response: %v", err)
		c.Redirect(http.StatusFound, "https://www.hyperlane.cc/login?error=invalid_token")
		return
	}

	// 获取用户信息
	accessToken := tokenResp.AccessToken
	reqArgs.URL = viper.GetString("oauth.getUser")
	reqArgs.Method = "GET"
	header = make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", accessToken)
	reqArgs.Headers = header

	userResult, err := utils.SendHTTPRequest(reqArgs)
	if err != nil {
		logger.Log.Errorf("Get user info error: %v", err)
		c.Redirect(http.StatusFound, "https://www.hyperlane.cc/login?error=user_info_failed")
		return
	}

	var githubUser GitHubUserResponse
	err = json.Unmarshal([]byte(userResult), &githubUser)
	if err != nil || githubUser.ID == 0 {
		logger.Log.Errorf("Invalid user data: %v", err)
		c.Redirect(http.StatusFound, "https://www.hyperlane.cc/login?error=invalid_user")
		return
	}

	// 创建或更新用户
	var user *models.User
	userId := uint(githubUser.ID)
	user, err = models.GetUserByUid(userId)

	if err == nil {
		// 更新已存在的用户
		user.Uid = userId
		user.Email = githubUser.Email
		user.Github = githubUser.HTMLURL
		err = models.UpdateUser(user)
	} else {
		// 创建新用户
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
		logger.Log.Errorf("Save user error: %v", err)
		c.Redirect(http.StatusFound, "https://www.hyperlane.cc/login?error=save_user_failed")
		return
	}

	// 获取用户权限
	perms, err := models.GetUserWithPermissions(user.ID)
	if err != nil {
		logger.Log.Errorf("Get permissions error: %v", err)
		c.Redirect(http.StatusFound, "https://www.hyperlane.cc/login?error=permissions_failed")
		return
	}

	// 生成 JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, user.Avatar, user.Username, user.Github, perms)
	if err != nil {
		logger.Log.Errorf("Generate token error: %v", err)
		c.Redirect(http.StatusFound, "https://www.hyperlane.cc/login?error=token_failed")
		return
	}

	// 成功：重定向到首页并带上 token
	c.Redirect(http.StatusFound, fmt.Sprintf("https://www.hyperlane.cc/?token=%s", token))
}
