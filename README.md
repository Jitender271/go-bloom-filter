# Scalable Bloom Filter in Go

This project implements a Scalable Bloom Filter in Go, allowing efficient membership testing for large datasets while minimizing false positive rates. The filter automatically adjusts its capacity and false positive rates based on the number of inserted elements.

## Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Concurrency](#concurrency)
- [License](#license)

## Introduction

A Bloom Filter is a probabilistic data structure that provides a way to test whether an element is a member of a set. It can yield false positives but never false negatives. This implementation extends the standard Bloom Filter to be scalable, dynamically adjusting its size and false positive rate as more elements are added.

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/yourusername/go-bloom-filter.git
   cd go-bloom-filter
   go build
   go run main.go -config=config.json
   OR
   go run main.go -defaults=true
   ```

## Usage
Run the application:

You can run the application with a custom configuration or use the default configuration.

Example Configuration:

A sample config.json file may look like this:

```bash
{
  "initial_fp": 0.01,
  "growth_factor": 2.0,
  "tightening_ratio": 0.5,
  "initial_capacity": 1000
}

Add Elements:

The application adds elements to the Bloom Filter and checks for membership:


elementsToAdd := []string{"apple", "banana", "cherry"}
for _, item := range elementsToAdd {
    err := sbf.Add(item)
}```

Check Membership:

To check if an element might be present:


contains := sbf.MightContain("apple")
```

## Configuration

initial_fp: Initial false positive rate (should be between 0 and 1).

growth_factor: The factor by which capacity grows (should be greater than 1).

tightening_ratio: Ratio to reduce the false positive rate (should be between 0 and 1).

initial_capacity: Initial expected number of elements (should be greater than 0).

## Concurrency
This implementation is designed to be concurrent-safe. It uses mutexes to handle read and write operations, ensuring that multiple goroutines can interact with the Bloom Filter without causing data races.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.
