package engine

type engineUpdate struct {
	id     int
	update func(float64)
}

type Updater struct {
	updates    map[int]engineUpdate
	backAdd    []engineUpdate
	backRemove []int
	nextId     int
	lastDelta  float64
	pending    chan int
	complete   chan int
}

func NewUpdater() Updater {
	return Updater{
		updates:    make(map[int]engineUpdate),
		backAdd:    make([]engineUpdate, 0),
		backRemove: make([]int, 0),
		nextId:     1,
		pending:    make(chan int, 100),
		complete:   make(chan int, 100),
	}
}

func (u *Updater) StartThreads(threads int) {
	for i := 0; i < threads; i++ {
		go u.updateThread()
	}
}

func (u *Updater) updateThread() {
	// TODO:  Does this need to be cleaned up?
	for {
		id := <-u.pending
		u.updates[id].update(u.lastDelta)
		u.complete <- id
	}
}

func (u *Updater) addInternal() {
	for _, update := range u.backAdd {
		u.updates[update.id] = update
	}
	u.backAdd = u.backAdd[:0]
}

func (u *Updater) removeInternal() {
	for _, id := range u.backRemove {
		delete(u.updates, id)
	}
	u.backRemove = u.backRemove[:0]
}

func (u *Updater) AddUpdate(update func(float64)) int {
	id := u.nextId
	u.backAdd = append(u.backAdd, engineUpdate{
		id:     id,
		update: update,
	})
	u.nextId++
	return id
}

func (u *Updater) RemoveUpdate(id int) {
	if id > 0 {
		u.backRemove = append(u.backRemove, id)
	}
}

func (u *Updater) inlineUpdate(deltaTime float64) {
	for _, eu := range u.updates {
		eu.update(deltaTime)
	}
}

func (u *Updater) threadedUpdate() {
	waitCount := 0
	for id := range u.updates {
		waitCount++
		u.pending <- id
	}
	for i := 0; i < waitCount; i++ {
		<-u.complete
	}
}

func (u *Updater) Update(deltaTime float64) {
	u.lastDelta = deltaTime
	u.addInternal()
	u.removeInternal()
	u.inlineUpdate(deltaTime)
	//u.threadedUpdate()
}
