package peer

import (
	"sort"
	"sync"
	"time"
)

type PresenceManager struct {
	mutex                 sync.Mutex
	localNickname         string
	activeUsers           map[string]time.Time
	timeoutDuration       time.Duration
	checkInterval         time.Duration
	onUserJoinedCallback  func(userNickname string)
	onUserTimeoutCallback func(userNickname string)
	onUserLeftCallback    func(userNickname string)
	stopChannel           chan struct{}
}

func NewPresenceManager(
	localNickname string,
	timeoutDuration time.Duration,
	checkInterval time.Duration,
	onUserJoinedCallback func(userNickname string),
	onUserTimeoutCallback func(userNickname string),
	onUserLeftCallback func(userNickname string),
) *PresenceManager {
	var presenceManager = &PresenceManager{
		localNickname:         localNickname,
		activeUsers:           make(map[string]time.Time),
		timeoutDuration:       timeoutDuration,
		checkInterval:         checkInterval,
		onUserJoinedCallback:  onUserJoinedCallback,
		onUserTimeoutCallback: onUserTimeoutCallback,
		onUserLeftCallback:    onUserLeftCallback,
		stopChannel:           make(chan struct{}),
	}

	go presenceManager.runExpirationLoop()

	return presenceManager
}

func (presenceManager *PresenceManager) RecordActivity(userNickname string) bool {
	if userNickname == "" || userNickname == presenceManager.localNickname {
		return false
	}

	presenceManager.mutex.Lock()
	defer presenceManager.mutex.Unlock()

	_, userExists := presenceManager.activeUsers[userNickname]
	presenceManager.activeUsers[userNickname] = time.Now()

	if !userExists {
		if presenceManager.onUserJoinedCallback != nil {
			go presenceManager.onUserJoinedCallback(userNickname)
		}
		return true
	}

	return false
}

func (presenceManager *PresenceManager) RemoveUser(userNickname string) bool {
	presenceManager.mutex.Lock()
	defer presenceManager.mutex.Unlock()

	_, userExists := presenceManager.activeUsers[userNickname]
	if userExists {
		delete(presenceManager.activeUsers, userNickname)
		if presenceManager.onUserLeftCallback != nil {
			go presenceManager.onUserLeftCallback(userNickname)
		}
		return true
	}

	return false
}

func (presenceManager *PresenceManager) GetActiveUsers() []string {
	presenceManager.mutex.Lock()
	defer presenceManager.mutex.Unlock()

	var userList = make([]string, 0, len(presenceManager.activeUsers)+1)
	userList = append(userList, presenceManager.localNickname)

	for activeNickname := range presenceManager.activeUsers {
		userList = append(userList, activeNickname)
	}

	sort.Strings(userList)
	return userList
}

func (presenceManager *PresenceManager) runExpirationLoop() {
	var ticker = time.NewTicker(presenceManager.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-presenceManager.stopChannel:
			return
		case currentTick := <-ticker.C:
			presenceManager.checkExpiredUsers(currentTick)
		}
	}
}

func (presenceManager *PresenceManager) checkExpiredUsers(currentTimestamp time.Time) {
	presenceManager.mutex.Lock()

	var expiredUsers []string
	for activeNickname, lastActivityTime := range presenceManager.activeUsers {
		if currentTimestamp.Sub(lastActivityTime) >= presenceManager.timeoutDuration {
			expiredUsers = append(expiredUsers, activeNickname)
		}
	}

	for _, expiredNickname := range expiredUsers {
		delete(presenceManager.activeUsers, expiredNickname)
	}

	presenceManager.mutex.Unlock()

	if presenceManager.onUserTimeoutCallback != nil {
		for _, expiredNickname := range expiredUsers {
			presenceManager.onUserTimeoutCallback(expiredNickname)
		}
	}
}

func (presenceManager *PresenceManager) Stop() {
	close(presenceManager.stopChannel)
}
