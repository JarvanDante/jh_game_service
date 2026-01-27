package util

//BuildUsername
/**
 * @desc：构建游戏用户名
 * @param username
 * @param siteCode
 * @return string
 * @author : Carson
 */
func BuildUsername(username, siteCode string) string {
	return siteCode + "_" + username
}
