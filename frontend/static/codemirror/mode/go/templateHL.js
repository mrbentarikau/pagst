  var funcMap = {
    // bitwise functions
    "bitwiseAnd":true,
    "bitwiseOr":true,
    "bitwiseNot":true,
    "bitwiseXor":true,
    "bitwiseClear":true,
    "bitwiseAndNot":true,
    "bitwiseLeftShift":true,
    "bitwiseRightShift":true,
    "shiftLeft":true,
    "shiftRight":true,
    
    // conversion functions
    "decodeStringToHex":true,
    "hexToDecimal":true,
    "str":true,
    "toByte":true,
    "toDuration":true,
    "toFloat":true,
    "toInt":true,
    "toInt64":true,
    "toInt64Base16":true,
    "toRune":true,
    "toSHA256":true,
    "toString":true,
    
    // math
    "abs":true,
    "add":true,
    "cbrt":true,
    "cos":true,
    "div":true,
    "divMod":true,
    "exp":true,
    "exp2":true,
    "fdiv":true,
    "log":true,
    "mathConst":true,
    "max":true,
    "min":true,
    "mod":true,
    "mult":true,
    "pow":true,
    "round":true,
    "roundCeil":true,
    "roundFloor":true,
    "roundEven":true,
    "sin":true,
    "sqrt":true,
    "sub":true,
    "tan":true,
    
    // misc
    "adjective":true,
    "cembed":true,
    "complexMessage":true,
    "complexMessageEdit":true,
    "createTicket":true,
    "cslice":true,
    "dict":true,
    "derefPointer":true,
    "getApplicationCommands":true,
    "humanizeThousands":true,
    "in":true,
    "inFold":true,
    "json":true,
    "jsonToSdict":true,
    "kindOf":true,
    "noun":true,
    "ordinalize":true,
    "randFloat":true,
    "randInt":true,
    "replaceEmojis":true,
    "roleAbove":true,
    "sdict":true,
    "seq":true,
    "shuffle":true,
    "structToSdict":true,
    "verb":true,
    
    // string manipulation
    "hasPrefix":true, "hasSuffix":true, "joinStr":true, "lower":true,
    "normalizeAccents":true, "normalizeConfusables":true,
    "print":true, "println":true, "printf":true, "slice":true, "split":true, "title":true,
    "trim":true, "trimLeft":true, "trimRight":true, "trimSpace":true,
    "upper":true, "urlescape":true, "urlunescape":true,
    
    // time functions
    "currentTime":true,
    "formatTime":true,
    "loadLocation":true,
    "parseTime":true,
    "snowflakeToTime":true,
    "timestampToTime":true,
    "newDate":true,
    "weekNumber":true,
    "humanizeDurationHours":true,
    "humanizeDurationMinutes":true,
    "humanizeDurationSeconds":true,
    "humanizeTimeSinceDays":true,
    
    // context functions
    "editMessage":true,
    "editMessageNoEscape":true,
    "execTemplate":true,
    "lastMessages":true,
    "pinMessage":true,
    "sendDM":true, 
    "sendMessage":true,
    "sendMessageRetID":true,
    "sendMessageNoEscape":true,
    "sendMessageNoEscapeRetID":true,
    "sendTargetDM":true,
    "sendTemplate":true,
    "sendTemplateDM":true,
    "unpinMessage":true,
    
    // Mentions
    "mentionEveryone":true,
    "mentionHere":true,
    "mentionRole":true,
    "mentionRoleName":true,
    "mentionRoleID":true,
    
    // permission funcs
    "getTargetPermissionsIn":true,
    "hasPermissions":true,
    "setMemberTimeout":true,
    "targetHasPermissions":true,
    
    // Varia
    "addMessageReactions":true,
    "addResponseReactions":true,
    "addReactions":true,
    "ccCounters":true,
    "currentUserCreated":true,
    "currentUserAgeHuman":true,
    "currentUserAgeMinutes":true,
    "deleteAllMessageReactions":true,
    "deleteMessage":true,
    "deleteMessageReaction":true,
    "deleteResponse":true,
    "deleteTrigger":true,
    "editChannelName":true,
    "editChannelTopic":true,
    "editNickname":true,
    "getAllMessageReactions":true,
    // "getAuditLogEntries":true,
    "getBotCount":true,
    "getChannel":true,
    "getChannelPins":true,
    "getChannelOrThread":true,
    "getMember":true,
    "getMemberTimezone":true,
    "getMessage":true,
    "getMessageReactions":true,
    "getPinCount":true,
    "getThread":true,
    "getUser":true,
    "getUserCount":true,
    "onlineCount":true,
    "onlineCountBots":true,
    "pastNicknames":true,
    "pastUsernames":true,
    "reFind":true,
    "reFindAll":true,
    "reFindAllSubmatches":true,
    "reQuoteMeta":true,
    "reReplace":true,
    "reSplit":true,
    "sleep":true,
    "sort":true,
    
    // templatextensions
    "cancelScheduledUniqueCC":true,
    "carg":true,
    "editCCTriggerType":true,
    "execCC":true,
    "parseArgs":true,
    "scheduleUniqueCC":true,
    
    // templexec
    "exec":true,
    "execAdmin":true,
    "userArg":true,

    // Role functions
    "addRole":true,
    "addRoleID":true,
    "addRoleName":true,
    "getRole":true,
    "getRoleID":true,
    "getRoleName":true,
    "giveRole":true,
    "giveRoleID":true,
    "giveRoleName":true,
    "hasRole":true,
    "hasRoleID":true,
    "hasRoleName":true,
    "removeRole":true,
    "removeRoleID":true,
    "removeRoleName":true,
    "setRoles":true,
    "takeRole":true,
    "takeRoleID":true,
    "takeRoleName":true,
    "targetHasRole":true,
    "targetHasRoleID":true,
    "targetHasRoleName":true,

// template user database
    "dbBottomEntries":true,
    "dbCount":true,
    "dbDecr":true,
    "dbDel":true,
    "dbDelByID":true,
    "dbDelById":true,
    "dbDelMultiple":true,
    "dbGet":true,
    "dbGetByID":true,
    "dbGetPattern":true,
    "dbIncr":true,
    "dbRank":true,
    "dbSet":true,
    "dbSetExpire":true,
    "dbGetPatternReverse":true,
    "dbTopEntries":true
  };
