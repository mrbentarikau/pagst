// local trivia uses some questions from Pixel Puzzles Trivia - SteamAppID: 1005210
package trivia

var LocalQuestions = map[string][]TriviaQuestion{
	"Entertainment - Movies": {
		{
			Answer:   "Sylvester Stallone",
			Question: "Who wrote the screenplay for Rocky?",
			Options:  []string{"Joel Coen", "James Cameron", "Alexander Payne"},
		},
		{
			Answer:   "White Castle",
			Question: "In the 2004 comedy starring John Cho and Kal Penn, Harold and Kumar go where?",
			Options:  []string{"Wendy's", "Burger King", "McDonald's"},
		},
		{
			Answer:   "Math",
			Question: "In the film Good Will Hunting, what educational subject does Will specialise in?",
			Options:  []string{"Geography", "History", "Arts"},
		},
		{
			Answer:   "David Bowie",
			Question: "In the film Zoolander, who judgest the walk off between Ben Stiller and Owen Wilson's characters?",
			Options:  []string{"Will Ferrell", "Christian Slater", "Gwen Stefani"},
		},
		{
			Answer:   "Ponyo",
			Question: "Which Studio Ghibli movie die Liam Neeson lend his voice?",
			Options:  []string{"Spirited Away", "The Secret World of Arrietty", "The Wind Rises"},
		},
	},
	"Entertainment - Video Games": {
		{
			Answer:   "Playing Cards",
			Question: "What was Nintendo's first product?",
			Options:  []string{"Nintendo Entertainment System", "Game & Watch", "Table Soccer"},
		},
		{
			Answer:   "Pac-Man",
			Question: "What game entered the Guinness Book of Records as the most successful coin-operated game of all-time?",
			Options:  []string{"Space Invaders", "Golden Axe", "Street Fighter II"},
		},
		{
			Answer:   "Atari",
			Question: "Steve Jobs once worked for which video game company?",
			Options:  []string{"Coleco", "Sega", "Nintendo"},
		},
		{
			Answer:   "Gil",
			Question: "What is the currency used in the 'Final Fantasy' franchise?",
			Options:  []string{"Coins", "Gold", "Crystals"},
		},
		{
			Answer:   "E.T. the Extra-Terrestrial",
			Question: "Which game was famously discarded in a New Mexico landfill due to being a massive commercial failure?",
			Options:  []string{"Shaq Fu", "Bubsy 3D", "Dr. Jekyll and Mr. Hyde"},
		},
	},
	"Science - Space": {
		{
			Answer:   "Isaac Newton",
			Question: "Which famous scientist and mathematician formulated the laws of motion and universal gravitation, laying the foundation for modern physics?",
			Options:  []string{"Johannes Kepler", "Galileo Galilei", "Gottfried Wilhelm Leibniz"},
		},
		{
			Answer:   "John Glenn",
			Question: "Who was the first American to orbit the Earth, completing three orbits in the Friendship 7 spacecraft?",
			Options:  []string{"Scott Carpenter", "Alan Shepard", "Gus Grissom"},
		},
		{
			Answer:   "Ursa Minor",
			Question: "Of what constellation is the Little Dipper a part?",
			Options:  []string{"Draco", "Lyra", "Cepheus"},
		},
		{
			Answer:   "Hayabusa",
			Question: "Which mission was the first to return samples from an asteroid, collecting material from the surface of Itokawa in 2005?",
			Options:  []string{"Rosetta", "OSIRIS-REx", "Dawn"},
		},
		{
			Answer:   "Galileo",
			Question: "What is the name of the spacecraft that successfully orbited Jupiter and its moons, providing crucial data about the gas giant and its Galilean moons?",
			Options:  []string{"Juno", "Cassini", "New Horizons"},
		},
	},
	"Art": {
		{
			Answer:   "Antonio Canova",
			Question: "Which Neoclassical Italian sculptor is celebrated for his marble sculptures, often portraying classical themes and figures, like 'Psyche Revived by Cupid's Kiss' and 'The Three Graces'?",
			Options:  []string{"Auguste Rodin", "Donatello", "Gian Lorenzo Bernini"},
		},
		{
			Answer:   "Salvador Dalí",
			Question: "Who painted 'The Persistence of Memory' featuring melting clocks?",
			Options:  []string{"Claude Monet", "Vincent van Gogh", "Pablo Picasso"},
		},
		{
			Answer:   "Georgia O'Keeffe",
			Question: "Which American artist is renowned for her paintings of enlarged flowers, New York skyscrapers, and New Mexico landscapes, often exploring the themes of sexuality and femininity?",
			Options:  []string{"Yayoi Kusama", "Tamara de Lempicka", "Frida Kahlo"},
		},
		{
			Answer:   "Mark Rothko",
			Question: "Which American abstract expressionist painter is recognized for his large colour-field paintings like 'No. 61 (Rust and Blue)' and 'White Center (Yellow, Pink and Lavender on Rose)'?",
			Options:  []string{"Jackson Pollock", "Franz Kline", "Willem de Kooning"},
		},
		{
			Answer:   "Leonardo da Vinci",
			Question: "Which artist painted the famous scene of 'The Last Supper'?",
			Options:  []string{"Claude Monet", "Michelangelo", "Caravaggio"},
		},
	},
}
