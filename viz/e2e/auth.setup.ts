import { test as setup, expect, type Page } from '@playwright/test';

const authFile = 'e2e/.auth/user.json';

async function handleOnboarding(page: Page) {
    console.log('Handling onboarding flow...');
    
    // Step 0: Welcome
    const getStartedBtn = page.locator('button:has-text("Get Started"), button:has-text("Let\'s Go")');
    if (await getStartedBtn.isVisible()) {
        await getStartedBtn.click();
    }

    // Steps 1..N: Settings & Details
    let maxLoops = 15;
    while (maxLoops--) {
        // Fill all visible text inputs for validation
        const inputs = page.locator('input[type="text"], input:not([type])');
        const count = await inputs.count();
        for (let i = 0; i < count; i++) {
            const input = inputs.nth(i);
            if (await input.isVisible() && await input.isEditable()) {
                const val = await input.inputValue();
                if (!val) {
                    await input.fill('E2E Test');
                }
            }
        }

        const finishBtn = page.locator('button:has-text("Finish Setup"), button:has-text("Complete Setup")');
        if (await finishBtn.isVisible()) {
            await finishBtn.click();
            break;
        }
        
        const nextBtn = page.locator('button:has-text("Next")');
        if (await nextBtn.isVisible()) {
            await nextBtn.click();
        } else {
            // Check if we are already done or redirecting
            await page.waitForTimeout(1000);
            if (!page.url().includes('/onboarding')) break;
        }
        await page.waitForTimeout(500); // Wait for transition
    }
}

setup('authenticate', async ({ page }) => {
  const email = process.env.TEST_EMAIL;
  const password = process.env.TEST_PASSWORD;
  const username = process.env.TEST_USERNAME;

  if (!email || !password || !username) {
    throw new Error('Missing E2E test credentials. Please set TEST_EMAIL, TEST_PASSWORD, and TEST_USERNAME environment variables.');
  }

  console.log(`Navigating to app...`);
  await page.goto('/');

  const workspace = page.locator('.viz-workspace');
  
  // Wait for initial load
  await page.waitForLoadState('networkidle');
  console.log(`Current URL: ${page.url()}`);

  if (await workspace.isVisible()) {
    console.log('Already authenticated.');
  } else if (page.url().includes('/onboarding')) {
    console.log('First run detected. Assuming first run or incomplete setup.');
    await handleOnboarding(page);
    await expect(workspace).toBeVisible({ timeout: 20000 });
  } else {
    console.log('Not authenticated. Attempting login...');
    await page.fill('#login-email', email);
    await page.fill('#login-password', password);
    await page.click('#login-submit');

    // Wait for success, error, or onboarding redirect
    await Promise.race([
        workspace.waitFor({ state: 'visible', timeout: 15000 }).catch(() => {}),
        page.waitForURL('**/onboarding', { timeout: 10000 }).catch(() => {}),
        page.locator('.viz-toast-error').waitFor({ state: 'visible', timeout: 5000 }).catch(() => {})
    ]);

    if (page.url().includes('/onboarding')) {
        await handleOnboarding(page);
    }

    // Give it a moment to settle after potential onboarding
    if (await workspace.isVisible({ timeout: 15000 })) {
      console.log('Authentication and onboarding successful.');
    } else {
      const errorToast = page.locator('.viz-toast-error');
      if (await errorToast.isVisible()) {
          const errorMsg = await errorToast.locator('.viz-toast-message').textContent();
          console.error(`Login failed with toast: ${errorMsg}`);
          
          if (errorMsg?.includes('User not found')) {
              console.log('User not found. Attempting registration...');
              await page.goto('/auth/register');
              await page.fill('#reg-email', email);
              await page.fill('#reg-name', username);
              await page.fill('#reg-password', password);
              await page.fill('#reg-password-confirm', password);
              await page.click('#reg-submit');
              
              await page.waitForURL('**/onboarding', { timeout: 10000 }).catch(() => {});
              if (page.url().includes('/onboarding')) {
                  await handleOnboarding(page);
              }
              
              await expect(workspace).toBeVisible({ timeout: 15000 });
              console.log('Registration and onboarding successful.');
          } else {
              await page.screenshot({ path: 'e2e-login-failure.png' });
              throw new Error(`Authentication failed: ${errorMsg}`);
          }
      } else {
          await page.screenshot({ path: 'e2e-unknown-failure.png' });
          throw new Error(`Authentication failed: Unknown state at ${page.url()}`);
      }
    }
  }

  // Save state
  await page.context().storageState({ path: authFile });
});
