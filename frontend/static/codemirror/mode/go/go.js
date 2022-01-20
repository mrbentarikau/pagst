// CodeMirror, copyright (c) by Marijn Haverbeke and others
// Distributed under an MIT license: https://codemirror.net/LICENSE

(function(mod) {
  if (typeof exports == "object" && typeof module == "object") // CommonJS
    mod(require("../../lib/codemirror"));
  else if (typeof define == "function" && define.amd) // AMD
    define(["../../lib/codemirror"], mod);
  else // Plain browser env
    mod(CodeMirror);
})(function(CodeMirror) {
"use strict";

CodeMirror.defineMode("go", function(config) {
  var indentUnit = config.indentUnit;
  var largeRE = /\x2e[A-Z]/;
  var idRE = /[a-z_A-Z0-9'\xa1-\uffff]/;

  var keywords = {
    "break":true, "case":true, "chan":true, "const":true, "continue":true,
    "default":true, "defer":true, "else":true, "fallthrough":true, "for":true,
    "func":true, "go":true, "goto":true, "if":true, "import":true,
    "interface":true, "map":true, "package":true, "range":true, "return":true,
    "select":true, "struct":true, "switch":true, "type":true, "var":true,
    "bool":true, "byte":true, "complex64":true, "complex128":true,
    "float32":true, "float64":true, "int8":true, "int16":true, "int32":true,
    "int64":true, "string":true, "uint8":true, "uint16":true, "uint32":true,
    "uint64":true, "int":true, "uint":true, "uintptr":true, "error": true,
    "rune":true,
    "block":true, "define":true, "end":true, "template":true, "with":true
  };

  var atoms = {
    "true":true, "false":true, "iota":true, "nil":true, "append":true,
    "cap":true, "close":true, "complex":true, "copy":true, "delete":true, "imag":true,
    "len":true, "make":true, "new":true, "panic":true, "print":true,
    "printf":true, "println":true, "real":true, "recover":true,
    "and":true, "call":true, "html":true, "index":true, "js":true,
    "len":true, "not":true, "or":true,"urlquery":true, "eq":true, "ne":true,
    "lt":true, "le":true, "gt":true, "ge":true
  };

  var funcMap = {
    // bitwise functions
    "bitwiseAnd":true, "bitwiseOr":true, "bitwiseNot":true,
    "bitwiseXor":true, "bitwiseClear":true, "bitwiseAndNot":true,
    "bitwiseLeftShift":true, "bitwiseRightShift":true,
    "shiftLeft":true, "shiftRight":true,
    // conversion functions
    "decodeStringToHex":true, "hexToDecimal":true, "str":true, "toByte":true,
    "toDuration":true, "toFloat":true, "toInt":true, "toInt64":true,
    "toInt64Base16":true, "toRune":true, "toString":true, "toSHA256":true,
    // math
    "abs":true, "add":true, "cbrt":true, "cos":true, "div":true,"divMod":true,
    "exp":true, "exp2":true, "fdiv":true, "log":true, "mod":true, "max":true,
    "min":true, "mult":true, "ordinalize": true, "pow":true, "round":true,
    "roundCeil":true, "roundFloor":true, "roundEven":true, "sin":true,
    "sqrt":true, "sub":true,"tan":true,
    // misc
    "adjective":true, "cembed":true, "complexMessage":true, "complexMessageEdit":true,
    "cslice":true, "dict":true, "humanizeThousands":true, "in":true, "inFold":true,
    "json":true, "kindOf":true, "noun":true, "randInt":true, "roleAbove":true,
    "sdict":true, "seq":true, "structToSdict":true, "shuffle":true,
    // string manipulation
    "hasPrefix":true, "hasSuffix":true, "joinStr":true, "lower":true,
    "print":true, "println":true, "printf":true, "slice":true, "split":true, "title":true,
    "trim":true, "trimLeft":true, "trimRight":true, "trimSpace":true,
    "upper":true, "urlescape":true, "urlunescape":true,
    // time functions
    "currentTime":true, "formatTime":true, "loadLocation":true,
    "parseTime":true, "snowflakeToTime":true, "newDate":true,
    "weekNumber":true,
    "humanizeDurationHours":true,
    "humanizeDurationMinutes":true,
    "humanizeDurationSeconds":true,
    "humanizeTimeSinceDays":true,
    // context functions
    "sendDM":true, 
    "sendTargetDM":true,
    "sendMessage":true,
    "sendTemplate":true,
    "sendTemplateDM":true,
    "sendMessageRetID":true,
    "sendMessageNoEscape":true,
    "sendMessageNoEscapeRetID":true,
    "editMessage":true,
    "editMessageNoEscape":true,
    "pinMessage":true,
    "unpinMessage":true,
    "lastMessages":true,
    // Mentions
    "mentionEveryone":true,
    "mentionHere":true,
    "mentionRole":true,
    "mentionRoleName":true,
    "mentionRoleID":true,
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
    // permission funcs
    "hasPermissions":true,
    "targetHasPermissions":true,
    "getTargetPermissionsIn":true,
    //Varia
    "deleteResponse":true,
    "deleteTrigger":true,
    "deleteMessage":true,
    "deleteMessageReaction":true,
    "deleteAllMessageReactions":true,
    "getMessage":true,
    "getAllMessageReactions":true,
    "getMember":true,
    "getChannel":true,
    "getThread":true,
    "getChannelOrThread":true,
    "getPinCount":true,
    "addReactions":true,
    "addResponseReactions":true,
    "addMessageReactions":true,
    "currentUserCreated":true,
    "currentUserAgeHuman":true,
    "currentUserAgeMinutes":true,
    "sleep":true,
    "reFind":true,
    "reFindAll":true,
    "reFindAllSubmatches":true,
    "reReplace":true,
    "reSplit":true,
    "editChannelTopic":true,
    "editChannelName":true,
    "onlineCount":true,
    "onlineCountBots":true,
    "editNickname":true,
    "sort":true,
    //templatextensions
    "cancelScheduledUniqueCC":true,
    "carg":true,
    "editCCTriggerType":true,
    "execCC":true,
    "parseArgs":true,
    "scheduleUniqueCC":true,
    //template user database
    "dbBottomEntries":true,
    "dbCount":true,
    "dbDecr":true,
    "dbDel":true,
    "dbDelByID":true,
    "dbDelById":true,
    "dbDelMultiple":true,
    "dbGet":true,
    "dbGetPattern":true,
    "dbIncr":true,
    "dbRank":true,
    "dbSet":true,
    "dbSetExpire":true,
    "dbGetPatternReverse":true,
    "dbTopEntries":true,
    //templexec
    "exec":true,
    "execAdmin":true,
    "userArg":true
  };

  var isOperatorChar = /[+\-*&^%:=<>!|\/]/;

  var curPunc;

  function tokenBase(stream, state) {
    var ch = stream.next();
    if (ch == '"' || ch == "'" || ch == "`") {
      state.tokenize = tokenString(ch);
      return state.tokenize(stream, state);
    }
    /*if (ch == "." && /[A-Z]/.test(stream.next())) {
      stream.eatWhile(idRE);
      //if (stream.eat('.')) {
      //  return "variable-3";
      //}
      return "variable-3";
    }*/
    if (/[\d\.]/.test(ch)) {
        if (ch == "." && stream.match(/[A-Z]/)) {
        stream.eatWhile(idRE);
        return "variable-3";
      } else if (ch == ".") {
        stream.match(/^[0-9]+([eE][\-+]?[0-9]+)?/);
      } else if (ch == "0") {
        stream.match(/^[xX][0-9a-fA-F]+/) || stream.match(/^0[0-7]+/);
      } else {
        stream.match(/^[0-9]*\.?[0-9]*([eE][\-+]?[0-9]+)?/);
      }
      return "number";
    }
    if (/[\[\]{}\(\),;\:\.]/.test(ch)) {
      curPunc = ch;
      return null;
    }
    if (ch == "/") {
      if (stream.eat("*")) {
        state.tokenize = tokenComment;
        return tokenComment(stream, state);
      }
      if (stream.eat("/")) {
        stream.skipToEnd();
        return "comment";
      }
    }
    if (isOperatorChar.test(ch)) {
      stream.eatWhile(isOperatorChar);
      return "operator";
    }
    stream.eatWhile(/[\w\$_\xa1-\uffff]/);
    var cur = stream.current();
    if (keywords.propertyIsEnumerable(cur)) {
      if (cur == "case" || cur == "default") curPunc = "case";
      return "keyword";
    }
    if (atoms.propertyIsEnumerable(cur)) return "atom";
    if (funcMap.propertyIsEnumerable(cur)) return "variable-3";
    return "variable";
  }

  function tokenString(quote) {
    return function(stream, state) {
      var escaped = false, next, end = false;
      while ((next = stream.next()) != null) {
        if (next == quote && !escaped) {end = true; break;}
        escaped = !escaped && quote != "`" && next == "\\";
      }
      if (end || !(escaped || quote == "`"))
        state.tokenize = tokenBase;
      return "string";
    };
  }

  function tokenComment(stream, state) {
    var maybeEnd = false, ch;
    while (ch = stream.next()) {
      if (ch == "/" && maybeEnd) {
        state.tokenize = tokenBase;
        break;
      }
      maybeEnd = (ch == "*");
    }
    return "comment";
  }

  function Context(indented, column, type, align, prev) {
    this.indented = indented;
    this.column = column;
    this.type = type;
    this.align = align;
    this.prev = prev;
  }
  function pushContext(state, col, type) {
    return state.context = new Context(state.indented, col, type, null, state.context);
  }
  function popContext(state) {
    if (!state.context.prev) return;
    var t = state.context.type;
    if (t == ")" || t == "]" || t == "}")
      state.indented = state.context.indented;
    return state.context = state.context.prev;
  }

  // Interface

  return {
    startState: function(basecolumn) {
      return {
        tokenize: null,
        context: new Context((basecolumn || 0) - indentUnit, 0, "top", false),
        indented: 0,
        startOfLine: true
      };
    },

    token: function(stream, state) {
      var ctx = state.context;
      if (stream.sol()) {
        if (ctx.align == null) ctx.align = false;
        state.indented = stream.indentation();
        state.startOfLine = true;
        if (ctx.type == "case") ctx.type = "}";
      }
      if (stream.eatSpace()) return null;
      curPunc = null;
      var style = (state.tokenize || tokenBase)(stream, state);
      if (style == "comment") return style;
      if (ctx.align == null) ctx.align = true;

      if (curPunc == "{") pushContext(state, stream.column(), "}");
      else if (curPunc == "[") pushContext(state, stream.column(), "]");
      else if (curPunc == "(") pushContext(state, stream.column(), ")");
      else if (curPunc == "case") ctx.type = "case";
      else if (curPunc == "}" && ctx.type == "}") popContext(state);
      else if (curPunc == ctx.type) popContext(state);
      state.startOfLine = false;
      return style;
    },

    indent: function(state, textAfter) {
      if (state.tokenize != tokenBase && state.tokenize != null) return CodeMirror.Pass;
      var ctx = state.context, firstChar = textAfter && textAfter.charAt(0);
      if (ctx.type == "case" && /^(?:case|default)\b/.test(textAfter)) {
        state.context.type = "}";
        return ctx.indented;
      }
      var closing = firstChar == ctx.type;
      if (ctx.align) return ctx.column + (closing ? 0 : 1);
      else return ctx.indented + (closing ? 0 : indentUnit);
    },

    electricChars: "{}):",
    closeBrackets: "()[]{}''\"\"``",
    fold: "brace",
    blockCommentStart: "/*",
    blockCommentEnd: "*/",
    lineComment: "//"
  };
});

CodeMirror.defineMIME("text/x-go", "go");

});
