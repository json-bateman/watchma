### Things to Do

- Next steps stuff: 

- services setup
    - [X] Setup logger, configure routing, make env variables accessable
    - [X] Figure out how to connect to jellyfin server
    - [X] Configure nats messaging system (removed, don't need the complexity for a single server)
    - [X] Room service to handle all of the apps rooms

- have room flow
    - [X] host page to host a room (done with form)
    - [X] table for all active rooms
    - [X] others can join room, assign users to room
    - [X] When host makes room, after leaving, doesn't remove him from room? Something is wrong with the
    join / leave logic (fixed)
    - [X] Make Ready Up! button work, user struct - ready: true? maybe not, but sure seems like a flag
    - [ ] Start Game! Button, should push users into a new endpoint maybe? /{room}/vote? Not sure
    how that would work, might have to just replace the entire screen via SSE?

- then game flow
    - [ ] move from lobby state to movies state
    - [ ] People vote on movies (/movies) (don't need SSE, everyone can have own instance)
    - [ ] Submit button to submit choices
    - [ ] View results!


- stretch goals
    - [ ] Generate a tournament bracket
    - [ ] People vote on the faceoffs 1 at a time
    - [ ] create DB to store results of finished games
    - [ ] let individual users log in
    - [ ] save users selections over time in the DB
