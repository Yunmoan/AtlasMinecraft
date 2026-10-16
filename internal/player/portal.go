package player

func (p *Player) PortalCooldown() int32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.portalCooldown
}

func (p *Player) SetPortalCooldown(ticks int32) {
	p.mu.Lock()
	p.portalCooldown = max(0, ticks)
	p.mu.Unlock()
}

// Each player has at most one contact check/transition in flight. Slow portal
// destination preparation cannot queue unbounded tick jobs or stall movement.
func (p *Player) BeginPortalCheck() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.portalChecking || !p.playReady {
		return false
	}
	p.portalChecking = true
	return true
}

func (p *Player) EndPortalCheck(transferred bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.portalChecking = false
	if transferred {
		p.portalCooldown, p.portalDuration = 10, 0
	}
}

// PortalProcessor compares the previous counter before incrementing: a
// creative Nether delay of 1 requires two contact ticks. Players renew a
// 10-tick cooldown while touching a portal, and count it down after leaving.
func (p *Player) AdvancePortalTick(inside bool, ticks, delay int32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.portalCooldown > 0 {
		if inside {
			p.portalCooldown = 9
		} else {
			p.portalCooldown = max(0, p.portalCooldown-ticks)
		}
		return false
	}
	if !inside {
		p.portalDuration = max(0, p.portalDuration-4*ticks)
		return false
	}
	if p.teleportPending {
		return false
	}
	p.portalDuration += ticks
	return p.portalDuration > delay
}
