package services

import (
	"fmt"
	"leaderboard-backend/models"
	"leaderboard-backend/store"
	"math/rand"

	"github.com/google/uuid"
)

type UserService struct {
	store       *store.MemoryStore
	ratingIndex *store.RatingBucketIndex
	minRating   int
	maxRating   int
}

func NewUserService(s *store.MemoryStore, ri *store.RatingBucketIndex, minRating, maxRating int) *UserService {
	return &UserService{
		store:       s,
		ratingIndex: ri,
		minRating:   minRating,
		maxRating:   maxRating,
	}
}

var firstNames = []string{
	"rahul", "priya", "arjun", "sneha", "vikram", "ananya", "amit", "neha",
	"raj", "pooja", "karan", "divya", "arun", "kavita", "suresh", "meera",
	"deepak", "nisha", "sandeep", "ritu", "ajay", "swati", "vijay", "anjali",
	"rohit", "varsha", "sanjay", "payal", "manish", "komal", "nikhil", "aarti",
	"sachin", "shruti", "rakesh", "preeti", "vishal", "jyoti", "gaurav", "smita",
	"harsh", "tanvi", "mohit", "shikha", "tushar", "rashmi", "varun", "megha",
	"ashish", "pallavi", "kapil", "sonali", "kunal", "kajal", "abhishek", "tanya",
	"pankaj", "garima", "ankit", "sakshi", "vikas", "monika", "akash", "dipti",
	"naveen", "archana", "dinesh", "namrata", "sumit", "richa", "tarun", "surbhi",
}

var lastNames = []string{
	"kumar", "sharma", "verma", "singh", "patel", "gupta", "joshi", "mehta",
	"reddy", "nair", "menon", "iyer", "rao", "pillai", "choudhary", "mishra",
	"agarwal", "banerjee", "chatterjee", "das", "mukherjee", "roy", "sen", "bose",
	"kapoor", "malhotra", "khanna", "arora", "sethi", "chopra", "bhatia", "kohli",
	"saxena", "mathur", "pandey", "tiwari", "dubey", "shukla", "tripathi", "srivastava",
	"burman", "jain", "shah", "thakur", "chauhan", "rajput", "yadav", "maurya",
}

func (u *UserService) GenerateUsername() string {
	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]

	format := rand.Intn(5)
	switch format {
	case 0:
		return firstName
	case 1:
		return fmt.Sprintf("%s_%s", firstName, lastName)
	case 2:
		return fmt.Sprintf("%s%d", firstName, rand.Intn(1000))
	case 3:
		return fmt.Sprintf("%s_%s%d", firstName, lastName, rand.Intn(100))
	default:
		return fmt.Sprintf("%s%s", firstName, lastName)
	}
}

func (u *UserService) GenerateRating() int {
	return u.minRating + rand.Intn(u.maxRating-u.minRating+1)
}

func (u *UserService) SeedUsers(count int) (int, error) {
	added := 0
	for i := 0; i < count; i++ {
		user := &models.User{
			ID:       uuid.New().String(),
			Username: u.GenerateUsername(),
			Rating:   u.GenerateRating(),
		}
		if err := u.store.AddUser(user); err == nil {
			added++
		}
	}
	return added, nil
}

func (u *UserService) UpdateRating(id string, newRating int) error {
	if newRating < u.minRating || newRating > u.maxRating {
		return fmt.Errorf("rating must be between %d and %d", u.minRating, u.maxRating)
	}
	return u.store.UpdateRating(id, newRating)
}

func (u *UserService) GetUser(id string) (*models.User, error) {
	return u.store.GetUser(id)
}

func (u *UserService) GetUserCount() int {
	return u.store.GetUserCount()
}

func (u *UserService) Clear() {
	u.store.Clear()
}
