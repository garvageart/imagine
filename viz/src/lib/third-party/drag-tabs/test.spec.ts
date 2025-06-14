import { test, expect, type Page } from '@playwright/test';

const TEST_MARKUP = `
  <div>
    <ul class="my-tabs-container">
      <li class="my-tab i-am-active" style="width: 30px; background: lightblue">A</li>
      <li class="my-tab" style="width: 15px; background: green">B</li>
      <li class="my-tab" style="width: 50px; background: yellow">C</li>
      <li class="my-tab" style="width: 80px; background: orange">D</li>
      <li class="my-tab" style="width: 50px; background: fuchsia">E</li>
      <li class="my-tab" style="width: 80px; background: lightgreen">F</li>
      <li class="my-tab" style="width: 80px; background: lime">G</li>
      <li class="my-tab" style="width: 80px; background: red">H</li>
      <li class="my-tab" style="width: 80px; background: aliceblue">I</li>
      <li class="my-tab" style="width: 80px; background: cadetblue">J</li>
      <li class="my-tab ignore-me" style="width: 80px; background: grey">IGNORE</li>
    </ul>
  </div>
`;

const TEST_CSS = `.my-tab { display: inline-block; text-align: center; margin-right: 5px; }`;

async function injectDragTabs(page: Page) {
  await page.addScriptTag({ path: require.resolve('./index.ts') });
}

async function setupDOM(page: Page) {
  await page.setContent('<!DOCTYPE html><html><head></head><body></body></html>');
  await page.addStyleTag({ content: TEST_CSS });
  await page.evaluate((markup: string) => {
    document.body.innerHTML = markup;
  }, TEST_MARKUP);
}

test.describe('dragTabs', () => {
  test.beforeEach(async ({ page }) => {
    await setupDOM(page);
    await injectDragTabs(page);
  });

  test('should create dragger', async ({ page }) => {
    const result = await page.evaluate(() => {
      // @ts-ignore
      return typeof window.dragTabs === 'function';
    });

    expect(result).toBe(true);
    // Actually create the dragger
    const draggerExists = await page.evaluate(() => {
      // @ts-ignore
      const node = document.querySelector('div');
      // @ts-ignore
      const dragger = window.dragTabs(node, {
        selectors: {
          tabsContainer: '.my-tabs-container',
          tab: '.my-tab',
          ignore: '.ignore-me',
          active: '.i-am-active',
        },
      });
      return !!dragger;
    });

    expect(draggerExists).toBe(true);
  });

  test('should create without drag preview', async ({ page }) => {
    const draggerExists = await page.evaluate(() => {
      // @ts-ignore
      const node = document.querySelector('div');
      // @ts-ignore
      const dragger = window.dragTabs(node, {
        selectors: {
          tabsContainer: '.my-tabs-container',
          tab: '.my-tab',
          ignore: '.ignore-me',
          active: '.i-am-active',
        },
        showPreview: false,
      });
      return !!dragger;
    });

    expect(draggerExists).toBe(true);
  });

  test('should act as singleton', async ({ page }) => {
    const isSingleton = await page.evaluate(() => {
      // @ts-ignore
      const node = document.querySelector('div');
      // @ts-ignore
      const dragger = window.dragTabs(node, {
        selectors: {
          tabsContainer: '.my-tabs-container',
          tab: '.my-tab',
          ignore: '.ignore-me',
          active: '.i-am-active',
        },
      });

      // @ts-ignore
      const cachedDragger = window.dragTabs(node, false);
      return dragger === cachedDragger;
    });
    expect(isSingleton).toBe(true);
  });

  test('should emit events', async ({ page }) => {
    const result = await page.evaluate(() => {
      // @ts-ignore
      const node = document.querySelector('div');
      // @ts-ignore
      const dragger = window.dragTabs(node, {
        selectors: {
          tabsContainer: '.my-tabs-container',
          tab: '.my-tab',
          ignore: '.ignore-me',
          active: '.i-am-active',
        },
      });
      let events: string[] = [];
      dragger.on('start', () => events.push('start'));
      dragger.on('drag', () => events.push('drag'));
      dragger.on('end', () => events.push('end'));
      dragger.on('cancel', () => events.push('cancel'));

      // Simulate event emission
      dragger.emit('start', { dragTab: node, initialIndex: 0 });
      dragger.emit('drag', { dragTab: node, newIndex: 1 });
      dragger.emit('end', { dragTab: node, newIndex: 1 });
      dragger.emit('cancel', { dragTab: node, newIndex: 0 });
      return events;
    });

    expect(result).toEqual(['start', 'drag', 'end', 'cancel']);
  });
});
