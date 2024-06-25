package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var nouns = []string{
	"ant", "ape", "asp", "auk", "bass", "bat", "bear", "bee", "bird", "boar",
	"bug", "calf", "carp", "cat", "chimp", "cod", "colt", "conch", "cow",
	"crab", "crane", "croc", "crow", "cub", "deer", "doe", "dog", "dolphin",
	"donkey", "dove", "duck", "eel", "elk", "fawn", "finch", "fish", "flea",
	"fly", "foal", "fowl", "fox", "frog", "gnat", "gnu", "goat", "goose",
	"grouse", "grub", "gull", "hare", "hawk", "hen", "hog", "horse", "hound",
	"jay", "kit", "kite", "koi", "lamb", "lark", "loon", "lynx", "mare",
	"midge", "mink", "mole", "moose", "moth", "mouse", "mule", "newt", "owl",
	"ox", "panda", "penguin", "perch", "pig", "pug", "quail", "ram", "rat",
	"ray", "robin", "roo", "rook", "seal", "shad", "shark", "sheep", "shoat",
	"shrew", "shrike", "shrimp", "skink", "skunk", "sloth", "slug", "smelt",
	"snail", "snake", "snipe", "sow", "sponge", "squid", "squirrel", "stag",
	"steed", "stoat", "stork", "swan", "tern", "toad", "trout", "turtle",
	"vole", "wasp", "whale", "wolf", "worm", "wren", "yak", "zebra",
}

var predicates = []string{
	"abundant", "able", "abrasive", "adorable", "adaptable", "adventurous",
	"aged", "agreeable", "ambitious", "amazing", "amusing", "angry",
	"auspicious", "awesome", "bald", "beautiful", "bemused", "bedecked", "big",
	"bittersweet", "blushing", "bold", "bouncy", "brawny", "bright", "burly",
	"bustling", "calm", "capable", "carefree", "capricious", "caring",
	"casual", "charming", "chill", "classy", "clean", "clumsy", "colorful",
	"crawling", "dapper", "debonair", "dashing", "defiant", "delicate",
	"delightful", "dazzling", "efficient", "enchanting", "entertaining",
	"enthused", "exultant", "fearless", "flawless", "fortunate", "fun",
	"funny", "gaudy", "gentle", "gifted", "glamorous", "grandiose",
	"gregarious", "handsome", "hilarious", "honorable", "illustrious",
	"incongruous", "indecisive", "industrious", "intelligent", "inquisitive",
	"intrigued", "invincible", "judicious", "kindly", "languid", "learned",
	"legendary", "likeable", "loud", "luminous", "luxuriant", "lyrical",
	"magnificent", "marvelous", "masked", "melodic", "merciful", "mercurial",
	"monumental", "mysterious", "nebulous", "nervous", "nimble", "nosy",
	"omniscient", "orderly", "overjoyed", "peaceful", "painted", "persistent",
	"placid", "polite", "popular", "powerful", "puzzled", "rambunctious",
	"rare", "rebellious", "respected", "resilient", "righteous", "receptive",
	"redolent", "resilient", "rogue", "rumbling", "salty", "sassy", "secretive",
	"selective", "sedate", "serious", "shivering", "skillful", "sincere",
	"skittish", "silent", "smiling",
}

const numRange = 1000

// GenerateRandomName generates random name for `run`.
func GenerateRandomName() (string, error) {
	predicateIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(predicates))))
	if err != nil {
		return "", fmt.Errorf("error getting random integer number: %w", err)
	}

	nounIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	if err != nil {
		return "", fmt.Errorf("error getting random integer number: %w", err)
	}

	num, err := rand.Int(rand.Reader, big.NewInt(numRange))
	if err != nil {
		return "", fmt.Errorf("error getting random integer number: %w", err)
	}

	return fmt.Sprintf("%s-%s-%d", predicates[predicateIndex.Int64()], nouns[nounIndex.Int64()], num), nil
}
