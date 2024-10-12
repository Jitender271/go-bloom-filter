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
