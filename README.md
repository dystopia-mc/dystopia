
# Dystopia
MC:BE Practice server written in golang using [dragonfly](https://github.com/df-mc/dragonfly).

I've abandoned it because it has very annoying deadlock, that happens so rarely, that I'd better rewrite entire project and make it open-source, than debugging this.

I was thinking that this issue with my `ExecSafe` function:

```go
func (pl *Player) ExecSafe(f func(*player.Player, *world.Tx)) {
	go pl.Player.H().ExecWorld(func(tx *world.Tx, e world.Entity) {
		f(e.(*player.Player), tx)
	})
}
```

But I was wrong. It happens only with >=3 online, when player is teleporting to the lobby instance 🤡

I'm 100% sure that this is an issue with new dragonfly transactions, but I didn't want to use old dragonfly version, so I kept working with latest commits. Hope in the nearest future transactions will be a little bit less usable than now.

# What's inside?
There is so many stupid and not smart solutions, so get ready trying to read code. 
Also, almost no comments, because I was lazy to document a lot of code. So hope you'll understand what I've written.

# Good parts of this project

Settings are working, pots and pearls are good configured, so you can steal config to use at your project.
I also used unique solutions, like using packets `SetHud` to manage hud, `UpdateClientInputLocks` to lock player input and other solutions that I've only seen on my project.

Pearl Cooldown and combat mode working perfectly. All possible ways, i.e. exit in the middle of the game or /suicide are provided.

# Credits
FPS counter:
* [divinity.adiavi.com](https://divinity.adiavi.com) 
* [swimgg.club](https://swimgg.club)

Combat mode, knockback, pearls, pots configurations:
* [antralia.com](https://antralia.com) (I relied on their system)
