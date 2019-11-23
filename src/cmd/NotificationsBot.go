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
    "strconv"
    "math/rand"
    "time"
    "strings"
    "github.com/benkillin/NotificationsBot/src/EmbedHelper"
    "io/ioutil"
)

var (
    configFile = "BotConfig.json"
    defaultConfigFile = "BotConfig.default.json" // this file gets overwritten every run with the current default config
    botID string // Bot ID
    config *Config
    breadMap map[string]BreadEntry
    breadEntries int
)

// our main function
func main() {
    rand.Seed(time.Now().UTC().UnixNano())

    defaultConfig := &Config{
        Token: "",
        Logging: LoggingConfig {
            Level: "trace",
            Format: "text",
            Output: "stderr",
            Logfile: ""},
        Guilds: map[string]*GuildConfig{
            "123456789012345678": &GuildConfig{
                GuildName: "DerpGuild",
                CommandPrefix: ".",
                KeywordsEnabled: false,
                Players: map[string]*PlayerConfig{
                    "123456789012345678": &PlayerConfig{
                        PlayerString: "Derp#1234",
                        PlayerUsername: "asdfasdfasdf",
                        PlayerMention: "@123456789012345678",
                        Keywords: []string{"key1", "key2"},
                        KeywordsEnabled: false}}}}} // the default config
    config = &Config{} // the running configuration

    // This is debug code basically to keep the default json file updated which is checked into git.
    os.Remove(defaultConfigFile)
    ConfigHelper.GetConfigWithDefault(defaultConfigFile, defaultConfig, &Config{})
    
    err := ConfigHelper.GetConfigWithDefault(configFile, defaultConfig, config)
	if err != nil {
		log.Fatalf("error loading/saving config/default config. %s", err)
    }
    
    setupLogging(config)

    // load bread
    log.Debugf("Loading bread...")
    breadText, err := ioutil.ReadFile("bread.txt")
    if err != nil {
        log.Fatalf("Error loading bread: %s", err)
    }
    breadLines := strings.Split(strings.Replace(string(breadText), "\r", "", -1), "\n")
    breadEntries = len(breadLines)
    breadMap = make(map[string]BreadEntry)
    for _, line := range breadLines {
        breadColumns := strings.Split(line, "\t")

        breadMap[breadColumns[0]] = BreadEntry{
            Name: breadColumns[1],
            Type: breadColumns[2],
            Description: breadColumns[3],
        }
    }
    log.Debugf("End loading bread...")
    // end loading bread

    token := config.Token
    
	d, err := discordgo.New("Bot " + token)
    if err != nil {
        log.Fatalf("Failed to create discord session: %s", err)
    }
    log.Infof("Created discord object.")

    bot, err := d.User("@me")
    if err != nil {
        log.Fatalf("Failed to get the bot user/access account: %s", err)
    }
    log.Infof("Obtained self user.")

	botID = bot.ID
    d.AddHandler(messageHandler)

    err = d.Open()
    if err != nil {
        log.Fatalf("Error: unable to establish connection to discord: %s", err)
    }
    log.Infof("Successfully opened discord connection.")

    defer d.Close()

    <-make(chan struct{})
}


// our command handler function
func messageHandler(d *discordgo.Session, msg *discordgo.MessageCreate) {
    if msg.GuildID == "" {
        return
    }
    
    checkGuild(d, msg.ChannelID, msg.GuildID)
    content := msg.Content
    splitContent := strings.Split(content, " ")
    prefix := config.Guilds[msg.GuildID].CommandPrefix
    if strings.HasPrefix(splitContent[0], prefix) {
        switch splitContent[0]{
        case prefix + "test":
            testCmd(d, msg.ChannelID, msg, splitContent)
        case prefix + "set":
            setCmd(d, msg.ChannelID, msg, splitContent)
        case prefix + "keyword":
            keywordCmd(d, msg.ChannelID, msg, splitContent)
        case prefix + "help":
            helpCmd(d, msg.ChannelID, msg, splitContent, availableCommands)
        case prefix + "bread":
            breadCmd(d, msg.ChannelID, msg, splitContent)
        case prefix + "invite":
            deleteMsg(d, msg.ChannelID, msg.ID)
            ch, err := d.UserChannelCreate(msg.Author.ID)
            if err != nil {
                errmsg := fmt.Sprintf("Error creating user channel for private message with invite link: %s", err)
                log.Error(errmsg)
                sendTempMsg(d, msg.ChannelID, errmsg, 60*time.Second)
                break
            }
            sendMsg(d, ch.ID, fmt.Sprintf("Here is a link to invite this bot to your own server: https://discordapp.com/api/oauth2/authorize?client_id=%s&permissions=8&scope=bot", botID))
        case prefix + "lennyface":
            deleteMsg(d, msg.ChannelID, msg.ID)
            sendMsg(d, msg.ChannelID, "( ͡° ͜ʖ ͡°)")
        case prefix + "tableflip":
            fallthrough
        case prefix + "fliptable":
            deleteMsg(d, msg.ChannelID, msg.ID)
            sendMsg(d, msg.ChannelID, "(╯ ͠° ͟ʖ ͡°)╯┻━┻")
        case prefix + "grr":
            deleteMsg(d, msg.ChannelID, msg.ID)
            sendMsg(d, msg.ChannelID, "ಠ_ಠ")
        case prefix + "manylenny":
            fallthrough
        case prefix + "manyface":
            deleteMsg(d, msg.ChannelID, msg.ID)
            sendMsg(d, msg.ChannelID, "( ͡°( ͡° ͜ʖ( ͡° ͜ʖ ͡°)ʖ ͡°) ͡°)")
        case prefix + "finger":
            deleteMsg(d, msg.ChannelID, msg.ID)
            sendMsg(d, msg.ChannelID, "凸-_-凸")
        case prefix + "gimme":
            deleteMsg(d, msg.ChannelID, msg.ID)
            sendMsg(d, msg.ChannelID, "ლ(´ڡ`ლ)")
        case prefix + "shrug":
            deleteMsg(d, msg.ChannelID, msg.ID)
            sendMsg(d, msg.ChannelID, "¯\\_(ツ)_/¯")
        }
    } else {
        // loop through the keywords and send out notifications if required.
        for playerIndex, playerValue := range config.Guilds[msg.GuildID].Players {
            if playerValue.KeywordsEnabled {
                for _, keywordValue := range playerValue.Keywords {
                    if strings.Contains(msg.Content, keywordValue) {
                        log.Infof("got a keyword hit: %#v", msg.Content)
                        ch, err := d.UserChannelCreate(playerIndex)
                        if err != nil {
                            errmsg := fmt.Sprintf("Error creating user channel for private message with invite link: %s", err)
                            log.Error(errmsg)
                            break
                        }
                        guild, err := d.Guild(msg.GuildID)
                        if err != nil {
                            log.Errorf("unable to obtain guild while sending a notification")
                        }
                        incomingChannel, err := d.Channel(msg.ChannelID)
                        if err != nil {
                            log.Errorf("Unable to obtain channel while sending a notification")
                        }
                        embed := EmbedHelper.NewEmbed().
                        SetTitle("¡Keyword alert!").
                        AddField("Discord server", guild.Name).
                        AddField("Channel", incomingChannel.Mention()).
                        AddField("Keyword phrase", "`" + keywordValue + "`").
                        AddField("Sending user", msg.Author.Mention()).
                        MessageEmbed
                    
                        _, err = sendEmbed(d, ch.ID, embed)
                        if err != nil {
                            log.Errorf("Error sending embed in response to keyword: %#v %#v", embed, msg.Content)
                        }
                    } // end checking if the keyword was in the message.
                } // end looping over keywords
            } // end checking if the player had keyword alerts enabled
        } // end looping over players
    } // end checking if we are looking for a command
}

// Settings command - set the various settings that make the bot operate on a particular guild aka server.
func setCmd(d *discordgo.Session, channelID string, msg *discordgo.MessageCreate, splitMessage []string) {
    if len(splitMessage) > 1 {
        //deleteMsg(d, msg.ChannelID, msg.ID) // let's not delete settings commands in case someone does something nefarious.
        log.Debugf("Incoming settings message: %+v", msg.Message)

        checkGuild(d, channelID, msg.GuildID)
        err := checkRole(d, msg, config.Guilds[msg.GuildID].RoleAdmin)
        if err != nil {
            sendMsg(d, msg.ChannelID, fmt.Sprintf("NotificationsBot role check failed. Contact someone who can assign you the correct role for bot settings."))
            return
        }

        subcommand := splitMessage[1]

        switch subcommand {
        case "keywords":
            if len(splitMessage) > 2 {
                changed := false

                switch splitMessage[2] {
                case "on":
                    config.Guilds[msg.GuildID].KeywordsEnabled = true
                    changed = true
                    sendTempMsg(d, channelID, fmt.Sprintf("Keyword notifications are now enabled!"), 45 * time.Second)

                case "off":
                    config.Guilds[msg.GuildID].KeywordsEnabled = false
                    changed = true
                    sendTempMsg(d, channelID, fmt.Sprintf("Keyword notifications are now disabled."), 45 * time.Second)

                
                case "admin":
                    isAdmin, err := MemberHasPermission(d, msg.GuildID, msg.Author.ID, discordgo.PermissionAdministrator)
                    if err != nil {
                        log.Debugf("Unable to determine if user is admin: %s", err)
                        sendTempMsg(d, channelID, fmt.Sprintf("Error: Unable to determine user permissions: %s", err), 45*time.Second)
                    }

                    if isAdmin {
                        if len(msg.MentionRoles) > 0 {
                            admin := msg.MentionRoles[0]
                            config.Guilds[msg.GuildID].RoleAdmin = admin
                            changed = true
                            sendTempMsg(d, channelID, "Set bot admin to role <@&" + config.Guilds[msg.GuildID].RoleAdmin + ">", 60*time.Second)
                        } else {
                            sendTempMsg(d, channelID, "Error - invalid/no role specified", 60*time.Second)
                        }
                    } else {
                        sendMsg(d, channelID, "Error - only server/guild administrators may change this setting.")
                    }
                default:
                    sendCurrentBotSettings(d, channelID, msg)
                }

                if changed {
                    ConfigHelper.SaveConfig(configFile, config)
                    sendCurrentBotSettings(d, channelID, msg)
                }
            } else {
                sendCurrentBotSettings(d, channelID, msg)
            }

        case "prefix":
            if len(splitMessage) > 2 {
                prefix := splitMessage[2]
                config.Guilds[msg.GuildID].CommandPrefix = prefix
                ConfigHelper.SaveConfig(configFile, config)
            } else {
                sendTempMsg(d, channelID, "usage: " + config.Guilds[msg.GuildID].CommandPrefix + "set prefix {command prefix here. example: . or !! or ! or ¡ or ¿}", 10*time.Second)
            }
        default: 
            helpCmd(d, channelID, msg, splitMessage, setCommands)
        }
    } else {
        helpCmd(d, channelID, msg, splitMessage, setCommands)
    }
}

func keywordCmd(d *discordgo.Session, channelID string, msg *discordgo.MessageCreate, splitMessage []string) {
    user := msg.Author
    if user.ID == botID || user.Bot || msg.GuildID == "" {
        return
    }
    
    checkGuild(d, msg.ChannelID, msg.GuildID)
    _, err := checkPlayer(d, channelID, msg.GuildID, msg.Author.ID)
    if err != nil {
        log.Errorf("Unable to check the player. %s", err)
        return
    }

    changed := false

    if len(splitMessage) > 2 {
        switch splitMessage[1] {
        case "add":
            changed = true
            phrase := strings.Join(splitMessage[2:], " ")
            config.Guilds[msg.GuildID].Players[msg.Author.ID].Keywords = append(config.Guilds[msg.GuildID].Players[msg.Author.ID].Keywords, phrase)
            sendTempMsg(d, channelID, fmt.Sprintf("Added '%s' as a keyword to alert you when said.", phrase), 60*time.Second)
        case "remove":
            changed = true
            theBirds := config.Guilds[msg.GuildID].Players[msg.Author.ID].Keywords // the birds = the words... :D
            phrase := strings.Join(splitMessage[2:], " ")
            found := false
            for index, value := range theBirds {
                if value == phrase {
                    config.Guilds[msg.GuildID].Players[msg.Author.ID].Keywords = remove(theBirds, index)
                    sendTempMsg(d, channelID, fmt.Sprintf("Removed '%s' from keyword alerts.", phrase), 60*time.Second)
                    found = true
                    break
                }
            }
            if !found {
                sendTempMsg(d, channelID, fmt.Sprintf("Unable to find '%s' in your keywords.", phrase), 60*time.Second)
            }
        default:
            log.Errorf("Invalid keyword command specified: %#v", splitMessage)
            sendCurrentKeywordSettings(d, channelID, msg)
        }
    } else {
        if len(splitMessage) > 1 {
            switch splitMessage[1] {
            case "on":
                changed = true
                config.Guilds[msg.GuildID].Players[msg.Author.ID].KeywordsEnabled = true
                sendTempMsg(d, channelID, "Enabled individual keyword alerts.", 30*time.Second)
            case "off":
                changed = true
                config.Guilds[msg.GuildID].Players[msg.Author.ID].KeywordsEnabled = false
                sendTempMsg(d, channelID, "DISABLED individual keyword alerts.", 30*time.Second)
            }
        } else {
            sendCurrentKeywordSettings(d, channelID, msg)
        }
    }

    if changed {
        ConfigHelper.SaveConfig(configFile, config)
    }

}

// Help command - explains the different commands the bot offers.
func helpCmd(d *discordgo.Session, channelID string, msg *discordgo.MessageCreate, splitMessage []string, commands []CmdHelp) {
    deleteMsg(d, msg.ChannelID, msg.ID)

    embed := EmbedHelper.NewEmbed().SetTitle("Available commands").SetDescription("Below are the available commands")

    for _, command := range commands {
        embed = embed.AddField(config.Guilds[msg.GuildID].CommandPrefix + command.command, command.description)
    }

    sendEmbed(d, channelID, embed.MessageEmbed)
}

func testCmd(d *discordgo.Session, channelID string, msg *discordgo.MessageCreate, splitMessage []string) {
    log.Debugf("Incoming TEST Message: %+v\n", msg.Message)
    messageIds := make([]string, 0)
    log.Debugf("Mention of author: %s; String of author: %s; author ID: %s", msg.Author.Mention(), msg.Author.String(), msg.Author.ID)

    deleteMsg(d, msg.ChannelID, msg.ID)
    
    msgID := sendMsg(d, msg.ChannelID, fmt.Sprintf("Hello, %s, you have initated a test of the self destruct sequence!", msg.Author.Mention()))
    messageIds = append(messageIds, msgID)

    for i := 5; i > 0; i-- {
        msgID := sendMsg(d, msg.ChannelID, fmt.Sprintf("%d", i))
        messageIds = append(messageIds, msgID)
        time.Sleep(1500 * time.Millisecond) // it seems if it is 1 second or faster then discord itself will throttle.
    }

    time.Sleep(3 * time.Second)

    err := d.ChannelMessagesBulkDelete(msg.ChannelID, messageIds)
    if err != nil {
        log.Errorf("Error: Unable to delete messages: %s", err)
    }
}

func breadCmd(d *discordgo.Session, channelID string, msg *discordgo.MessageCreate, splitMessage []string) {
    log.Debugf("Incoming random bread request! Message: %+v\n", msg.Message)
    messageIds := make([]string, 0)
    deleteMsg(d, msg.ChannelID, msg.ID)
    
    msgID := sendMsg(d, msg.ChannelID, fmt.Sprintf("Hello, %s, you have summoned bread! Summoning bread...", msg.Author.Mention()))
    messageIds = append(messageIds, msgID)

    breadEmoticons := [...]string{":bread:", ":french_bread:", ":stuffed_flatbread:"}

    for i := 3; i > 0; i-- {
        msgID := sendMsg(d, msg.ChannelID, breadEmoticons[rand.Intn(3)])
        messageIds = append(messageIds, msgID)
        time.Sleep(1000 * time.Millisecond) // it seems if it is 1 second or faster then discord itself will throttle.
    }

    time.Sleep(1 * time.Second)

    err := d.ChannelMessagesBulkDelete(msg.ChannelID, messageIds)
    if err != nil {
        log.Errorf("Error: Unable to delete messages: %s", err)
    }

    randomBreadNumber := rand.Intn(breadEntries+1)
    bread := breadMap[strconv.FormatInt(int64(randomBreadNumber), 10)]
    log.Debugf("Random bread number: %s, bread: %#v", randomBreadNumber, bread)
    
    embed := EmbedHelper.NewEmbed().
    SetTitle("Random Bread!").
    AddField(fmt.Sprintf("Bread random roll (0-%d)", breadEntries), fmt.Sprintf("%d", randomBreadNumber)).
    AddField("Bread Name", bread.Name).
    AddField("Bread Type", bread.Type).
    AddField("Bread Description", bread.Description).
    MessageEmbed

    chID := msg.ChannelID

    if strings.Contains(msg.Content, "private") {
        ch, err := d.UserChannelCreate(msg.Author.ID)
        if err != nil {
            errmsg := fmt.Sprintf("Error creating user channel for private message with invite link: %s", err)
            log.Error(errmsg)
            sendTempMsg(d, msg.ChannelID, errmsg, 60*time.Second)
            return
        }

        chID = ch.ID
        sendMsg(d, msg.ChannelID, "Sent random bread result in DM.")
    }

    _, err = sendEmbed(d, chID, embed)
    if err != nil {
        log.Errorf("Error sending embed in response to bread: %#v %#v", embed, msg.Content)
    }
}