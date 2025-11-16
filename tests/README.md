# Watchma E2E Tests

End-to-end tests for Watchma using Playwright.

## Setup

Tests are already configured! Just make sure you have dependencies installed:

```bash
# Install Node dependencies (includes Playwright)
npm install

# Install Playwright browsers (already done during npm install above)
npx playwright install chromium
```

## Running Tests

### Run all tests (headless)
```bash
npm test
```

### Watch tests run in browser
```bash
npm run test:headed
```

### Debug tests with Playwright Inspector
```bash
npm run test:debug
```

### Interactive UI mode (best for development)
```bash
npm run test:ui
```

### View last test report
```bash
npm run test:report
```

## Test Structure

```
tests/e2e/
├── helpers/
│   ├── auth.ts       # Authentication helper functions
│   ├── rooms.ts      # Room management helpers
│   └── game.ts       # Game flow helpers
├── auth.spec.ts      # Authentication tests
└── game-flow.spec.ts # Full game cycle tests
```

## Writing Tests

### Example: Testing room creation

```typescript
import { test, expect } from '@playwright/test';
import { signup, generateUsername } from './helpers/auth';
import { createRoom, generateRoomName } from './helpers/rooms';

test('user can create room', async ({ page }) => {
  const username = generateUsername('host');
  const roomName = generateRoomName('MyRoom');

  await signup(page, username, 'password123');
  await createRoom(page, roomName, 3, 4);

  await expect(page).toHaveURL(`/room/${roomName}`);
});
```

### Testing multiple users

```typescript
test('multiple users can join room', async ({ browser }) => {
  // Create separate contexts for different users
  const userA = await browser.newContext();
  const userB = await browser.newContext();

  const pageA = await userA.newPage();
  const pageB = await userB.newPage();

  // ... test with both users ...

  await userA.close();
  await userB.close();
});
```

## Available Helpers

### Authentication (`helpers/auth.ts`)
- `signup(page, username, password)` - Sign up new user
- `login(page, username, password)` - Log in existing user
- `logout(page)` - Log out current user
- `generateUsername(prefix)` - Generate unique username
- `isAuthenticated(page)` - Check if user is logged in

### Rooms (`helpers/rooms.ts`)
- `createRoom(page, name, draftNumber, maxPlayers)` - Create new room
- `joinRoom(page, roomName)` - Join existing room
- `clickReady(page)` - Click ready button in lobby
- `startGame(page)` - Start game (host only)
- `waitForPlayer(page, username)` - Wait for player to appear
- `generateRoomName(prefix)` - Generate unique room name
- `leaveRoom(page)` - Leave current room

### Game Flow (`helpers/game.ts`)
- `selectMoviesInDraft(page, count)` - Select N movies during draft
- `submitDraft(page)` - Submit draft selections
- `voteForMovies(page, count)` - Vote for N movies
- `submitVote(page)` - Submit vote
- `waitForResults(page, timeout)` - Wait for results screen
- `waitForPhase(page, phase)` - Wait for specific game phase
- `searchMovies(page, query)` - Search for movies
- `filterByGenre(page, genre)` - Filter by genre
- `sortMovies(page, sortBy)` - Sort movies

## Configuration

Tests are configured in `playwright.config.ts`:

- **Base URL**: `http://localhost:58008`
- **Browser**: Chromium (can enable Firefox/WebKit in config)
- **Workers**: 1 (sequential tests due to SQLite)
- **Retries**: 2 in CI, 0 locally
- **Timeout**: 30 seconds per test
- **Auto-start server**: Yes (runs `go run cmd/main.go`)

## CI/CD

Tests run automatically on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`

GitHub Actions workflow: `.github/workflows/playwright.yml`

## Debugging Failed Tests

### View trace
When a test fails, Playwright captures a trace. View it with:
```bash
npm run test:report
```

Click on a failed test to see:
- Screenshots at each step
- Network requests
- Console logs
- Detailed timeline

### Run single test
```bash
npx playwright test auth.spec.ts
```

### Run specific test by name
```bash
npx playwright test -g "user can sign up"
```

### Generate code
Record interactions to generate test code:
```bash
npx playwright codegen http://localhost:58008
```

## Tips

1. **Use unique names**: Always use `generateUsername()` and `generateRoomName()` to avoid conflicts
2. **Wait for SSE**: Use `waitForPlayer()` or add small timeouts after SSE-triggered actions
3. **Clean up contexts**: Always close browser contexts in multi-user tests
4. **Check selectors**: Update selectors in helpers if your HTML structure changes
5. **Test data**: Tests use dummy movie data by default (set `USE_DUMMY_DATA=true`)

## Troubleshooting

### Tests timing out
- Increase timeout in `playwright.config.ts`
- Check if server started successfully
- Verify SSE connections aren't blocked

### Selectors not found
- Update selectors in helper files
- Add `data-testid` attributes to your Templ components
- Use Playwright Inspector to debug: `npm run test:debug`

### Database conflicts
- Tests run sequentially (workers: 1) to avoid SQLite conflicts
- Each test creates unique users/rooms to avoid collisions
- Database is reset between test runs (fresh `watchma.db`)

## Next Steps

1. Add more test scenarios (edge cases, error handling)
2. Add visual regression testing with Playwright screenshots
3. Test with different viewport sizes (mobile/tablet)
4. Add performance testing (measure page load times)
5. Add accessibility testing (a11y checks)
