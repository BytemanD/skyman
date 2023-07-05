// Openstack 认证客户端

package identity

import (
	"fmt"
	"os"
	"strconv"
)

const DEFAULT_TOKEN_EXPIRE_SECOND = 3600

func getTokenExpireSecond() int {
	osSecond := os.Getenv("OS_TOKEN_EXPIRE_SECOND")
	if osSecond == "" {
		return DEFAULT_TOKEN_EXPIRE_SECOND
	} else {
		second, err := strconv.Atoi(osSecond)
		if err == nil {
			return second
		} else {
			return DEFAULT_TOKEN_EXPIRE_SECOND
		}

	}
}

// 获取认证客户端
func GetV3Client(authUrl string, user map[string]string, project map[string]string, regionName string) (V3AuthClient, error) {
	if authUrl == "" {
		return V3AuthClient{}, fmt.Errorf("authUrl is missing")
	}

	client := V3AuthClient{
		AuthUrl:           authUrl,
		Username:          user["name"],
		Password:          user["password"],
		UserDomainName:    user["domainName"],
		ProjectName:       project["name"],
		ProjectDomainName: project["domainName"],
		RegionName:        regionName,
		TokenExpireSecond: getTokenExpireSecond(),
	}
	if client.RegionName == "" {
		client.RegionName = "RegionOne"
	}
	return client, nil
}
