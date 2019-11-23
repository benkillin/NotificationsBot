package main
// Copyright (c) 2019 Benkillin. 
// This program is distributed under the terms of the GNU Affero General Public License..
// See LICENSE for the full license.

import (
    //"time"
)

// Config represents the application's configuration
type Config struct {
    Token string
    Logging LoggingConfig
    Guilds map[string]*GuildConfig
}

// LoggingConfig configuration as part of the config object.
type LoggingConfig struct {
    Level string
    Format string
    Output string
    Logfile string
}

// GuildConfig represents the configuration of a single instance of this bot on a particular server/guild
type GuildConfig struct {
    GuildName string
    RoleAdmin string
    CommandPrefix string
    KeywordsEnabled bool
    Players map[string]*PlayerConfig
}

// PlayerConfig represents the players and their scores.
type PlayerConfig struct {
    PlayerString string
    PlayerUsername string
    PlayerMention string
    Keywords []string // list of keywords to alert on. 
    KeywordsEnabled bool
}

// CmdHelp represents a key value pair of a command and a description of a command for constructing a help message embed.
type CmdHelp struct {
    command string
    description string
}

// BreadEntry represents a name description tuple of a type of bread.
type BreadEntry struct {
    Name string
    Type string
    Description string
}

var availableCommands = []CmdHelp {CmdHelp {command: "test", description:"A test command."},
        CmdHelp {command: "keyword", description:"Manage personal keywords for alerts."},
        CmdHelp {command: "set", description:"Manage bot settings."},
        CmdHelp {command: "bread", description:"Randomly select a type of bread. If you add the word 'private' as a parameter, then the resulting random bread selection will be DM'd to you instead of put in the channel."},
        CmdHelp {command: "help", description:"This help command menu."},
        CmdHelp {command: "invite", description:"Private message you the invite link for this bot to join a server you are an administrator of."},
        CmdHelp {command: "lennyface", description:"Emoji: giggity"},
        CmdHelp {command: "fliptable", description:"Emoji: FLIP THE FREAKING TABLE"},
        CmdHelp {command: "grr", description:"Emoji: i am angry or disappointed with you"},
        CmdHelp {command: "manyface", description:"Emoji: there is nothing but lenny"},
        CmdHelp {command: "finger", description:"Emoji: f you, man"},
        CmdHelp {command: "gimme", description:"Emoji: gimme gimme gimme gimme"},
        CmdHelp {command: "shrug", description:"Emoji: shrug things off"}}

var setCommands = []CmdHelp{CmdHelp {command:"set keywords on", description:"Enable keyword alerts for the server."},
    CmdHelp {command: "set keywords off", description:"Disable all keyword alerts for the server (Default)."},
    CmdHelp {command: "set keywords admin (role)", description: "The role to require to update bot settings on this server (Server administrators always allowed)."},
    CmdHelp {command: "set prefix (prefix)", description: "Set the command prefix to the specified string. (Defaults to .)."}}

var keywordCommands = []CmdHelp{CmdHelp{command:"keyword add (keyword)", description:"Add a keyword for alerts."},
    CmdHelp {command: "keyword remove (keyword)", description:"Remove a keyword for alerts."},
    CmdHelp {command: "keyword on", description: "Enable keyword notifications."},
    CmdHelp {command: "keyword off", description: "Disable keyword notifications (Default)."}}
