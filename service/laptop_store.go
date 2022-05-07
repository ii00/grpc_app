package service

import (
	"errors"
	"fmt"
	"grpc_app/pb"
	"sync"

	"github.com/jinzhu/copier"
)

// ErrAlreadyExist is returned when a record with the same ID already exists in the store.
var ErrAlreadyExist = errors.New("record already exist")

// LaptopStore is an interface to store laptop.
type LaptopStore interface {
	// Save saves the laptop to the store.
	Save(laptop *pb.Laptop) error
	// Find finds a laptop by ID.
	Find(id string) (*pb.Laptop, error)
	// Search searches for laptops with filter, returns one by one via the found function.
	Search(filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

// InMemoryLaptopStore stores laptop in memory.
type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

// DBLaptopStore stores laptop in DB.
// THIS ONE FOR LATER!
type DBLaptopStore struct{}

// NewInMemoryLaptopStore returns a new InMemoryLaptopStore.
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

// Save saves the laptop to the store
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExist
	}

	// deep copy
	other, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	store.data[other.Id] = other
	return nil
}

// Find finds a laptop by ID
func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RLocker().Unlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	// deep copy
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}
	return deepCopy(laptop)
}

// Search searches for laptops with filter, returns one by one via the found function.
func (store *InMemoryLaptopStore) Search(
	filter *pb.Filter,
	found func(laptop *pb.Laptop) error,

) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {
		if isQualified(filter, laptop) {
			// deep copy
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}

			err = found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 // 8 = 2^3 shift
	case pb.Memory_KILOBYTE:
		return value << 13 // 1024 * 8 = 2^10 * 2^3 = 2*13 shift
	case pb.Memory_MEGABYTE:
		return value << 23 // shift
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

// Deep copy.
func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return other, nil
}