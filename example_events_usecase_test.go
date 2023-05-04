package event_test

import (
	"fmt"
	event "go-event"
)

type Balances struct {
	list            map[uint64]*Balance
	events          event.Events[*Balance]
	newBalanceEvent event.Event[uint64]
}

func NewBalances() *Balances {
	return &Balances{
		list: make(map[uint64]*Balance),
	}
}

func (b *Balances) OnNewBalance(onNewBalance func(balanceId uint64)) {
	b.newBalanceEvent.On(onNewBalance)
}

func (b *Balances) OnBalanceChange(onBalanceChange func(*Balance)) {
	b.events.On("balance-change", onBalanceChange)
}

func (b *Balances) OnBalanceBlock(onBalanceBlock func(*Balance)) {
	b.events.On("balance-block", onBalanceBlock)
}

func (b *Balances) GetBalance(id uint64) *Balance {
	balance, ok := b.list[id]

	if ok {
		return balance
	}

	balance = &Balance{
		id:       id,
		balances: b,
	}

	b.list[id] = balance

	b.newBalanceEvent.Trigger(id)

	return balance
}

type Balance struct {
	id       uint64
	blocked  bool
	value    int64
	balances *Balances
}

func (b *Balance) Change(amount int64) {
	b.value += amount

	b.balances.events.Trigger("balance-change", b)
}

func (b *Balance) Block() {
	if !b.blocked {
		b.blocked = true

		b.balances.events.Trigger("balance-block", b)
	}
}

type Storage struct {
	data map[interface{}]interface{}
}

func (s *Storage) SaveValue(key, value interface{}) {
	s.data[key] = value

	fmt.Printf("Balance [%d] saved: %d\n", key, value)
}

/*
*
In this example we have a Balances struct representing user balances, it can manage balances and their values but can`t save them somewhere. But
we need to save balance values in any storage for restoring them in crash or app restart case. Also, we have a Storage struct representing a storage
(for example file storage). Events and Event help us to support SRP principe (Balances only know how to manage balances, Storage knows how to store any values) and
split technical logic of balances saving and business logic where we work with balances and do not think about lower level logic
*/
func ExampleEvents() {
	balances := NewBalances()

	//Set up a storage for balances values
	{
		storage := &Storage{data: make(map[interface{}]interface{})}

		balances.OnNewBalance(func(balanceId uint64) {
			fmt.Printf("Balance with id [%d] created\n", balanceId)
		})

		balances.OnBalanceChange(func(balance *Balance) {
			storage.SaveValue(balance.id, balance.value)
		})

		balances.OnBalanceBlock(func(balance *Balance) {
			fmt.Printf("Balance [%d] blocked with amount [%d]\n", balance.id, balance.value)
		})
	}

	//Business logic
	{
		balance := balances.GetBalance(1)

		balance.Change(10)
		balance.Change(-5)

		balance.Block()
	}

	// Output:
	// Balance with id [1] created
	// Balance [1] saved: 10
	// Balance [1] saved: 5
	// Balance [1] blocked with amount [5]
}
