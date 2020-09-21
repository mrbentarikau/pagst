package openweathermap

//parts copied from https://github.com/chubin/wttr.in/blob/master/share/we-lang/we-lang.go

var (
	windDirSlice = []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	windDir      = map[string]string{
		"N":   "↓",
		"NNE": "↓",
		"NE":  "↙",
		"ENE": "↙",
		"E":   "←",
		"ESE": "←",
		"SE":  "↖",
		"SSE": "↖",
		"S":   "↑",
		"SSW": "↑",
		"SW":  "↗",
		"WSW": "↗",
		"W":   "→",
		"WNW": "→",
		"NW":  "↘",
		"NNW": "↘",
	}

	iconUnknown = []string{
		"    .-.      ",
		"     __)     ",
		"    (        ",
		"     `-’     ",
		"      •      "}
	iconSunny = []string{
		"    \\   /    ",
		"     .-.     ",
		"  ― (   ) ―  ",
		"     `-’     ",
		"    /   \\    "}
	iconPartlyCloudy = []string{
		"   \\  /      ",
		" _ /\"\".-.    ",
		"   \\_(   ).  ",
		"   /(___(__) ",
		"             "}
	iconCloudy = []string{
		"             ",
		"     .--.    ",
		"  .-(    ).  ",
		" (___.__)__) ",
		"             "}
	iconVeryCloudy = []string{
		"             ",
		"     .--.    ",
		"  .-(    ).  ",
		" (___.__)__) ",
		"             "}
	iconLightShowers = []string{
		" _`/\"\".-.    ",
		"  ,\\_(   ).  ",
		"   /(___(__) ",
		"     ‘ ‘ ‘ ‘ ",
		"    ‘ ‘ ‘ ‘  "}
	iconHeavyShowers = []string{
		" _`/\"\".-.    ",
		"  ,\\_(   ).  ",
		"   /(___(__) ",
		"   ‚‘‚‘‚‘‚‘  ",
		"   ‚’‚’‚’‚’  "}
	iconLightSnowShowers = []string{
		" _`/\"\".-.    ",
		"  ,\\_(   ).  ",
		"   /(___(__) ",
		"     *  *  * ",
		"    *  *  *  "}
	iconHeavySnowShowers = []string{
		" _`/\"\".-.    ",
		"  ,\\_(   ).  ",
		"   /(___(__) ",
		"    * * * *  ",
		"   * * * *   "}
	iconLightSleetShowers = []string{
		" _`/\"\".-.    ",
		"  ,\\_(   ).  ",
		"   /(___(__) ",
		"     ‘ * ‘ * ",
		"    * ‘ * ‘  "}
	iconThunderyShowers = []string{
		" _`/\"\".-.    ",
		"  ,\\_(   ).  ",
		"   /(___(__) ",
		"    ⚡‘‘⚡‘‘ ",
		"    ‘ ‘ ‘ ‘  "}
	iconThunderyHeavyRain = []string{
		"     .-.     ",
		"    (   ).   ",
		"   (___(__)  ",
		"  ‚‘⚡‘‚⚡‚‘ ",
		"  ‚’‚’⚡’‚’  "}
	iconThunderySnowShowers = []string{
		" _`/\"\".-.    ",
		"  ,\\_(   ).  ",
		"   /(___(__) ",
		"     *⚡*⚡* ",
		"    *  *  *  "}
	iconLightRain = []string{
		"     .-.     ",
		"    (   ).   ",
		"   (___(__)  ",
		"    ‘ ‘ ‘ ‘  ",
		"   ‘ ‘ ‘ ‘   "}
	iconHeavyRain = []string{
		"     .-.     ",
		"    (   ).   ",
		"   (___(__)  ",
		"  ‚‘‚‘‚‘‚‘   ",
		"  ‚’‚’‚’‚’   "}
	iconLightSnow = []string{
		"     .-.     ",
		"    (   ).   ",
		"   (___(__)  ",
		"    *  *  *  ",
		"   *  *  *   "}
	iconHeavySnow = []string{
		"     .-.     ",
		"    (   ).   ",
		"   (___(__)  ",
		"   * * * *   ",
		"  * * * *    "}
	iconLightSleet = []string{
		"     .-.     ",
		"    (   ).   ",
		"   (___(__)  ",
		"    ‘ * ‘ *  ",
		"   * ‘ * ‘   "}
	iconFog = []string{
		"             ",
		" _ - _ - _ - ",
		"  _ - _ - _  ",
		" _ - _ - _ - ",
		"             "}
)
