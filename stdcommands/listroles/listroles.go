package listroles

import (
	"fmt"
	"math"
	"strings"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/lib/dstate"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"

	"github.com/mrbentarikau/pagst/common"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryTool,
	Name:        "ListRoles",
	Aliases:     []string{"lr", "ur"},
	Description: "List roles, their id's, color hex code, and 'mention everyone' perms (useful if you wanna double check to make sure you didn't give anyone mention everyone perms that shouldn't have it)",
	Arguments: []*dcmd.ArgDef{
		{Name: "User", Type: &commands.MemberArg{}},
	},

	ArgSwitches: []*dcmd.ArgDef{
		{Name: "nomanaged", Help: "Don't list managed/bot roles"},
		{Name: "managed", Help: "List managed/bot roles"},
		{Name: "raw", Help: "Raw, legacy output"},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var out, outFinal string
		var noMana, yesMana, raw bool
		var member *dstate.MemberState
		maxLength := 24

		if data.Args[0].Value != nil {
			member = data.Args[0].Value.(*dstate.MemberState)
		}

		if data.Switches["nomanaged"].Value != nil && data.Switches["nomanaged"].Value.(bool) && member == nil {
			noMana = true
		}

		if data.Switches["managed"].Value != nil && data.Switches["managed"].Value.(bool) && member == nil {
			yesMana = true
		}

		if data.Switches["raw"].Value != nil && data.Switches["raw"].Value.(bool) {
			raw = true
		}

		//sort.Sort(dutil.Roles(data.GuildData.GS.Roles))

		counter := 0
		if member != nil {
			if len(member.Member.Roles) > 0 {
				for _, roleID := range member.Member.Roles {
					for _, r := range data.GuildData.GS.Roles {
						if roleID == r.ID {
							counter++
							toOut(&r, raw, &out)
						}
					}
				}
			} else {
				return "User has no roles", nil
			}
		} else {
			for _, r := range data.GuildData.GS.Roles {
				if noMana && r.Managed {
					continue
				} else if yesMana && !r.Managed {
					continue
				} else {
					counter++
					toOut(&r, raw, &out)
				}
			}
		}

		if raw {
			outFinal = fmt.Sprintf("Total role count: %d\n", counter)
			outFinal += "(ME = mention everyone perms)\n"
			outFinal += out

			return outFinal, nil

		}
		//outSlice := strings.Split((strings.Replace(out, "`", "", -1)), "\n")
		outSlice := strings.Split(out, "\n")
		pm, err := paginatedmessages.CreatePaginatedMessage(
			data.GuildData.GS.ID, data.ChannelID, 1, int(math.Ceil(float64(len(outSlice))/float64(maxLength))), func(p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {
				i := page - 1
				paginatedEmbed := embedCreator(outSlice, i, maxLength, counter)
				return paginatedEmbed, nil
			})
		if err != nil {
			return fmt.Sprintf("Something went wrong: %s", err), nil
		}
		return pm, nil
	},
}

func toOut(r *discordgo.Role, raw bool, out *string) string {
	me := r.Permissions&discordgo.PermissionAdministrator != 0 || r.Permissions&discordgo.PermissionMentionEveryone != 0
	if !raw {
		nameCut := r.Name
		if common.ContainsEmoji(nameCut) {
			nameCut = common.ReplaceEmojis(nameCut, ":emoji:")
		}
		nameCut = common.CutStringShort(nameCut, 12)

		*out += fmt.Sprintf("%-12s %-19d#%-6x %-5t\n", nameCut, r.ID, r.Color, me)
	} else {
		*out += fmt.Sprintf("`%-25s: %-19d #%-6x  ME:%5t`\n", r.Name, r.ID, r.Color, me)
	}
	return ""
}

func embedCreator(outStringSlice []string, i, ml, counter int) *discordgo.MessageEmbed {
	description := fmt.Sprintf("Total role count: %d\n(ME = mention everyone perms)\n\n", counter)
	description += fmt.Sprintf("%-28s %-3s%-6s  %-5s\n", "Rolename", "ID", "Color", "ME")
	description += "---------------------------------------------\n"

	for k, v := range outStringSlice[i*ml:] {
		if k < ml {
			description += v + "\n"
		}
	}

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("`%s`", description),
	}
	return embed
}
