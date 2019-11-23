package main
// Copyright (c) 2019 Benkillin. 
// This program is distributed under the terms of the GNU Affero General Public License..
// See LICENSE for the full license.

import (
    "os"
    "fmt"
    "github.com/bwmarrin/discordgo"
    "github.com/benkillin/ConfigHelper"
    log "github.com/sirupsen/logrus"
    "time"
    //"strings"
    "github.com/benkillin/NotificationsBot/src/EmbedHelper"
)


func checkGuild(d *discordgo.Session, channelID string, GuildID string) (*discordgo.Guild, error) {
    guild, err := d.Guild(GuildID)
    if err != nil {
        log.Errorf("Error obtaining guild: %s", err)
        sendMsg(d, channelID, fmt.Sprintf("Error obtaining guild: %s", err))
        return nil, err
    }

    if _, ok := config.Guilds[GuildID]; !ok {
        players := make(map[string]*PlayerConfig)
        config.Guilds[GuildID] = &GuildConfig{
            GuildName: guild.Name,
            RoleAdmin: "",
            KeywordsEnabled: false,
            CommandPrefix: ".",
            Players: players}
    } else {
        if guild.Name != config.Guilds[GuildID].GuildName {
            config.Guilds[GuildID].GuildName = guild.Name
        } 
    }

    ConfigHelper.SaveConfig(configFile, config)

    return guild, nil
}

func checkPlayer(d *discordgo.Session, channelID string, GuildID string, authorID string) (*discordgo.User, error) {
    checkGuild(d, channelID, GuildID)
    player, err := d.User(authorID)
    if err != nil {
        log.Errorf("Error obtaining user information: %s", err)
        sendMsg(d, channelID, fmt.Sprintf("Error obtaining user information: %s", err))
        return nil, err
    }

    if _, ok := config.Guilds[GuildID].Players[player.ID]; !ok {
        config.Guilds[GuildID].Players[player.ID] = &PlayerConfig {
            PlayerString: player.String(),
            PlayerUsername: player.Username,
            PlayerMention: player.Mention(),
            Keywords: []string{},
            KeywordsEnabled: false}
    } else {
        if player.Username != config.Guilds[GuildID].Players[authorID].PlayerString {
            config.Guilds[GuildID].Players[authorID].PlayerString = player.String()
            config.Guilds[GuildID].Players[authorID].PlayerUsername = player.Username
            config.Guilds[GuildID].Players[authorID].PlayerMention = player.Mention()
        }
    }

    return player, nil
}

// check to see if the user is in the specified role, or is an administrator.
func checkRole(d *discordgo.Session, msg *discordgo.MessageCreate, requiredRole string) (error) {
    member, err := d.GuildMember(msg.GuildID, msg.Author.ID)
    if err != nil {
        log.Errorf("Error obtaining user information: %s", err)
        sendMsg(d, msg.ChannelID, fmt.Sprintf("Error obtaining user information: %s", err))
        return err
    }

    for _, role := range member.Roles {
        if role == requiredRole {
            log.Debugf("User passed role check.")
            return nil
        }
    }

    isAdmin, err := MemberHasPermission(d, msg.GuildID, msg.Author.ID, discordgo.PermissionAdministrator)
    if err != nil {
        log.Debugf("Unable to determine if user is admin: %s", err)
        return err
    }

    if isAdmin {
        log.Debugf("User passed role check (user is administrator).")
        return nil
    }

    log.Errorf("User %s <%s (%s)> does not have the correct role.", msg.Author.Username, member.Nick, msg.Author.Mention())
    return fmt.Errorf("user %s (%s) does not have the necessary role %s", msg.Author.Mention(), msg.Author.ID, requiredRole)
}

// send the current walls settings to the specified channel.
func sendCurrentBotSettings(d *discordgo.Session, channelID string, msg *discordgo.MessageCreate) {
    embed := EmbedHelper.NewEmbed().
        SetTitle("Bot settings").
        SetDescription("Current bot settings").
        AddField("Guild Name", config.Guilds[msg.GuildID].GuildName).
        AddField("Keyword notifications enabled", fmt.Sprintf("%t", config.Guilds[msg.GuildID].KeywordsEnabled)).
        AddField("Bot admin role", "<@&" + config.Guilds[msg.GuildID].RoleAdmin + ">").
        MessageEmbed
	
	sendTempEmbed(d, channelID, embed, 30*time.Second)
}

func sendCurrentKeywordSettings(d *discordgo.Session, channelID string, msg *discordgo.MessageCreate) {
    embed := EmbedHelper.NewEmbed().
        SetTitle("Individual keyword notification settings").
        SetDescription("Current notification settings").
        AddField("Player", config.Guilds[msg.GuildID].Players[msg.Author.ID].PlayerMention).
        AddField("Notification phrases", fmt.Sprintf("%#v", config.Guilds[msg.GuildID].Players[msg.Author.ID].Keywords)).
        AddField("Notifications enabled", fmt.Sprintf("%t", config.Guilds[msg.GuildID].Players[msg.Author.ID].KeywordsEnabled)).
        MessageEmbed
	
	sendTempEmbed(d, channelID, embed, 30*time.Second)
}

// helper func to send an embed message, aka a message that has a bunch of key value pairs and other things like images and stuff.
func sendEmbed(d *discordgo.Session, channelID string, embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
    msg, err := d.ChannelMessageSendEmbed(channelID, embed)

    if err != nil {
        log.Errorf("Error sending embed message: %s", err)
        return nil, err
    }

    return msg, nil
}

// sends an embed message and waits the specified duration in a separate goroutine prior to deleting the message.
func sendTempEmbed(d *discordgo.Session, channelID string, embed *discordgo.MessageEmbed, duration time.Duration) (*discordgo.Message, error) {
	msg, err := d.ChannelMessageSendEmbed(channelID, embed)

    if err != nil {
        log.Errorf("Error sending temp embed message: %s", err)
        return nil, err
	}
	
	go func() {
        time.Sleep(duration)
        deleteMsg(d, channelID, msg.ID)
    }()

    return msg, nil
}

// test func for the unit tests - can be removed if we can figure out how to do unit testing with the discord api mocked somehow.
func hello() (string) {
	return "Hello, world!"
}

// send a message including a typing notification.
func sendMsg(d *discordgo.Session, channelID string, msg string) (string) {
    err := d.ChannelTyping(channelID)
    if err != nil {
        log.Errorf("Unable to send typing notification: %s", err)
    }

    sentMessage, err := d.ChannelMessageSend(channelID, msg)
    if err != nil {
        log.Errorf("Unable to send message: %s", err)
        return ""
    }

    return sentMessage.ID
}

// delete a message
func deleteMsg(d *discordgo.Session, channelID string, messageID string) (error) {
    err := d.ChannelMessageDelete(channelID, messageID)
    if err != nil {
        log.Errorf("Error: Unable to delete incoming message: %s", err)
        return err
    }

    return nil
}

// send a self deleting message "this message will self destruct in 5..." :)
func sendTempMsg(d *discordgo.Session, channelID string, msg string, timeout time.Duration) {
    go func() {
        messageID := sendMsg(d, channelID, msg)
        time.Sleep(timeout)
        d.ChannelMessageDelete(channelID, messageID)
    }()
}

// set up the logger.
func setupLogging(config *Config) {

    if config.Logging.Format == "text" {
        log.SetFormatter(&log.TextFormatter{})
    } else if config.Logging.Format == "json" {
        log.SetFormatter(&log.JSONFormatter{})
    } else {
        log.Warning("Warning: unknown logging format specified. Allowed options are 'text' and 'json' for config.Logging.Format")
        log.SetFormatter(&log.TextFormatter{})
    }
	
    level, err := log.ParseLevel(config.Logging.Level)
    if err != nil {
        log.Fatalf("Error setting up logging - invalid parse level: %s", err)
    }

    log.SetLevel(level)

    if config.Logging.Output == "file" {
        file, err := os.OpenFile(config.Logging.Logfile, os.O_RDWR, 0644)
        if err != nil {
            log.Fatalf("Error opening log file: %s", err)
        }

        log.SetOutput(file)
    } else if config.Logging.Output == "stdout" {
        log.SetOutput(os.Stdout) // by default the package outputs to stderr
    } else if config.Logging.Output == "stderr" {
        // do nothing
    } else {
        log.Warn("Warning: log output option not recognized. Valid options are 'file' 'stdout' 'stderr' for config.Logging.output")
    }
}

// remove an element from a string array.
func remove(s []string, i int) []string {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}

// MemberHasPermission checks if a member has the given permission
// for example, If you would like to check if user has the administrator
// permission you would use
// --- MemberHasPermission(s, guildID, userID, discordgo.PermissionAdministrator)
// If you want to check for multiple permissions you would use the bitwise OR
// operator to pack more bits in. (e.g): PermissionAdministrator|PermissionAddReactions
// =================================================================================
//     s          :  discordgo session
//     guildID    :  guildID of the member you wish to check the roles of
//     userID     :  userID of the member you wish to retrieve
//     permission :  the permission you wish to check for
// from https://github.com/bwmarrin/discordgo/wiki/FAQ#permissions-and-roles
func MemberHasPermission(s *discordgo.Session, guildID string, userID string, permission int) (bool, error) {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.GuildMember(guildID, userID); err != nil {
			return false, err
		}
	}

    // Iterate through the role IDs stored in member.Roles
    // to check permissions
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			return false, err
		}
		if role.Permissions&permission != 0 {
			return true, nil
		}
	}

	return false, nil
}
