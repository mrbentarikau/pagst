package bot

import (
	"github.com/mrbentarikau/pagst/bot/shardmemberfetcher"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dstate"
)

var (
	memberFetcher *shardmemberfetcher.Manager
)

// GetMember will either return a member from state or fetch one from the member fetcher and then put it in state
// variadic options bool is used to get full member data for custom command functions userArg and getMember
func GetMember(guildID, userID int64, options ...bool) (*dstate.MemberState, error) {
	var ccExec bool
	if len(options) > 0 {
		ccExec = options[0]
	}

	if memberFetcher != nil && !ccExec {
		return memberFetcher.GetMember(guildID, userID)
	}

	ms := State.GetMember(guildID, userID)
	if ms != nil && ms.Member != nil && !ccExec {
		return ms, nil
	}

	member, err := common.BotSession.GuildMember(guildID, userID)
	if err != nil {
		return nil, err
	}

	return dstate.MemberStateFromMember(member), nil
}

// GetMembers is the same as GetMember but with multiple members
func GetMembers(guildID int64, userIDs ...int64) ([]*dstate.MemberState, error) {
	if memberFetcher != nil {
		return memberFetcher.GetMembers(guildID, userIDs...)
	}

	// fall back to something really slow
	result := make([]*dstate.MemberState, 0, len(userIDs))
	for _, v := range userIDs {
		r, err := GetMember(guildID, v)
		if err != nil {
			continue
		}

		result = append(result, r)
	}

	return result, nil
}
