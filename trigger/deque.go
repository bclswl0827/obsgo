package trigger

type deque struct {
	items []int
}

func (d *deque) extend(elements []int) {
	d.items = append(d.items, elements...)
}

func (d *deque) append(item int) {
	d.items = append(d.items, item)
}

func (d *deque) popLeft() (any, bool) {
	if len(d.items) == 0 {
		return nil, false
	}
	item := d.items[0]
	d.items = d.items[1:]
	return item, true
}

func (d *deque) appendLeft(item int) {
	d.items = append([]int{item}, d.items...)
}
