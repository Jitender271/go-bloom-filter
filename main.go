package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sync"
)

// Config holds the configuration parameters for the Scalable Bloom Filter.
type Config struct {
	InitialFP       float64 `json:"initial_fp"`       // Initial false positive rate
	GrowthFactor    float64 `json:"growth_factor"`    // Factor by which capacity grows
	TighteningRatio float64 `json:"tightening_ratio"` // Ratio to reduce false positive rate
	InitialCapacity int     `json:"initial_capacity"` // Initial expected number of elements
}

// ScalableBloomFilter represents a scalable bloom filter.
type ScalableBloomFilter struct {
	filters         []*BloomFilter
	initialFP       float64
	growthFactor    float64
	tighteningRatio float64
	initialCapacity int
	mutex           sync.RWMutex
}

// NewScalableBloomFilter creates a new ScalableBloomFilter with the given configuration.
// It validates the parameters to ensure they are within acceptable ranges.
func NewScalableBloomFilter(config Config) (*ScalableBloomFilter, error) {
	// Parameter Validation
	if config.TighteningRatio <= 0 || config.TighteningRatio >= 1 {
		return nil, errors.New("tighteningRatio must be between 0 and 1")
	}
	if config.GrowthFactor <= 1 {
		return nil, errors.New("growthFactor must be greater than 1")
	}
	if config.InitialFP <= 0 || config.InitialFP >= 1 {
		return nil, errors.New("initialFP must be between 0 and 1")
	}
	if config.InitialCapacity <= 0 {
		return nil, errors.New("initialCapacity must be greater than 0")
	}

	return &ScalableBloomFilter{
		filters:         []*BloomFilter{},
		initialFP:       config.InitialFP,
		growthFactor:    config.GrowthFactor,
		tighteningRatio: config.TighteningRatio,
		initialCapacity: config.InitialCapacity,
	}, nil
}

// Add inserts an item into the Scalable Bloom Filter.
// If the current Bloom filter cannot accommodate the new item, a new Bloom filter is created.
func (sbf *ScalableBloomFilter) Add(item string) error {
	sbf.mutex.Lock()
	defer sbf.mutex.Unlock()

	// If there are no filters or the last filter cannot accommodate the item, create a new filter
	if len(sbf.filters) == 0 || !sbf.filters[len(sbf.filters)-1].Add(item) {
		// Calculate new false positive probability using tighteningRatio
		newFP := sbf.initialFP * math.Pow(sbf.tighteningRatio, float64(len(sbf.filters)))

		// Calculate new capacity using growthFactor
		// Each new filter has capacity = initialCapacity * (growthFactor ^ number_of_filters)
		newCapacity := float64(sbf.initialCapacity) * math.Pow(sbf.growthFactor, float64(len(sbf.filters)))

		// Create a new Bloom filter with scaled capacity and adjusted false positive rate
		newFilter := NewBloomFilter(int(math.Ceil(newCapacity)), newFP)

		// Add the item to the new filter
		newFilter.Add(item)

		// Append the new filter to the list of filters
		sbf.filters = append(sbf.filters, newFilter)
	}
	return nil
}

// MightContain checks if an item might be in the Scalable Bloom Filter.
// Returns true if the item might be present, false if it is definitely not present.
func (sbf *ScalableBloomFilter) MightContain(item string) bool {
	sbf.mutex.RLock()
	defer sbf.mutex.RUnlock()

	for _, filter := range sbf.filters {
		if filter.MightContain(item) {
			return true
		}
	}
	return false
}

// BloomFilter represents a single Bloom filter.
type BloomFilter struct {
	bitset       []uint8
	bitSize      uint
	numHashFuncs uint
	mutex        sync.RWMutex
}

// NewBloomFilter creates a new BloomFilter with the given capacity and false positive probability.
func NewBloomFilter(n int, fp float64) *BloomFilter {
	m := optimalBitSize(n, fp)
	k := optimalHashFuncs(m, n)
	// Initialize the bitset with the number of bytes needed
	byteSize := (m + 7) / 8 // Round up to the nearest byte
	return &BloomFilter{
		bitset:       make([]uint8, byteSize),
		bitSize:      m,
		numHashFuncs: k,
	}
}

// Add inserts an item into the Bloom filter.
// Returns true if at least one bit was newly set (indicating a new item).
func (bf *BloomFilter) Add(item string) bool {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	hashes := bf.getHashes(item)
	isNew := false
	for _, hash := range hashes {
		byteIndex := hash / 8
		bitIndex := hash % 8
		if (bf.bitset[byteIndex] & (1 << bitIndex)) == 0 {
			isNew = true
			bf.bitset[byteIndex] |= (1 << bitIndex)
		}
	}
	return isNew
}

// MightContain checks if an item might be in the Bloom filter.
// Returns true if the item might be present, false if it is definitely not present.
func (bf *BloomFilter) MightContain(item string) bool {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	hashes := bf.getHashes(item)
	for _, hash := range hashes {
		byteIndex := hash / 8
		bitIndex := hash % 8
		if (bf.bitset[byteIndex] & (1 << bitIndex)) == 0 {
			return false
		}
	}
	return true
}

// getHashes generates the required number of hash indices for an item using double hashing.
func (bf *BloomFilter) getHashes(item string) []uint {
	data := md5.Sum([]byte(item))
	hash1 := binary.BigEndian.Uint32(data[0:4])
	hash2 := binary.BigEndian.Uint32(data[4:8])
	hashes := make([]uint, bf.numHashFuncs)
	for i := uint(0); i < bf.numHashFuncs; i++ {
		combinedHash := hash1 + uint32(i)*hash2
		hashes[i] = uint(combinedHash) % bf.bitSize
	}
	return hashes
}

// optimalBitSize calculates the optimal size of the bit array (m) for a Bloom filter.
func optimalBitSize(n int, p float64) uint {
	m := -float64(n) * math.Log(p) / (math.Pow(math.Log(2), 2))
	return uint(math.Ceil(m))
}

// optimalHashFuncs calculates the optimal number of hash functions (k) for a Bloom filter.
func optimalHashFuncs(m uint, n int) uint {
	k := (float64(m) / float64(n)) * math.Log(2)
	return uint(math.Round(k))
}

// loadConfig loads the configuration from a JSON file.
// Returns a Config struct or an error if loading fails.
func loadConfig(filepath string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func main() {
	// Define command-line flags for configuration
	configPath := flag.String("config", "config.json", "Path to configuration file")
	useDefaults := flag.Bool("defaults", false, "Use default configuration if true")
	flag.Parse()

	var config Config

	if *useDefaults {
		// Default configuration
		config = Config{
			InitialFP:       0.01, // 1% false positive rate
			GrowthFactor:    2.0,  // Capacity doubles with each new filter
			TighteningRatio: 0.5,  // False positive rate halves with each new filter
			InitialCapacity: 1000, // Initial expected number of elements
		}
	} else {
		// Load configuration from file
		if _, err := os.Stat(*configPath); os.IsNotExist(err) {
			fmt.Printf("Configuration file not found: %s\n", *configPath)
			os.Exit(1)
		}
		loadedConfig, err := loadConfig(*configPath)
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			os.Exit(1)
		}
		config = loadedConfig
	}

	// Initialize Scalable Bloom Filter with the loaded configuration
	sbf, err := NewScalableBloomFilter(config)
	if err != nil {
		fmt.Printf("Error initializing Scalable Bloom Filter: %v\n", err)
		os.Exit(1)
	}

	// Example usage: Add elements
	elementsToAdd := []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape"}

	for _, item := range elementsToAdd {
		err := sbf.Add(item)
		if err != nil {
			fmt.Printf("Error adding item '%s': %v\n", item, err)
		}
	}

	// Example usage: Check for elements
	elementsToCheck := []string{"apple", "banana", "cherry", "date", "kiwi", "lemon"}

	for _, item := range elementsToCheck {
		contains := sbf.MightContain(item)
		fmt.Printf("Contains '%s': %v\n", item, contains)
	}
}
