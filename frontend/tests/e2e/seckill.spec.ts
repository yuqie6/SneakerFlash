import { expect, test } from "@playwright/test"

test("product seckill pending flow redirects to order detail when ready", async ({ page }) => {
  await page.route("**/api/v1/login", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          access_token: "access-token",
          refresh_token: "refresh-token",
          expires_in: 3600,
        },
      }),
    })
  })

  await page.route("**/api/v1/profile", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          id: 1,
          username: "alice",
          avatar: "",
        },
      }),
    })
  })

  const productPayload = {
    id: 1,
    user_id: 1,
    name: "AJ 1 Chicago",
    price: 1299,
    stock: 8,
    start_time: "2024-01-01T00:00:00Z",
    image: "",
  }

  await page.route("**/api/v1/product/1", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: productPayload,
      }),
    })
  })

  await page.route("**/api/v1/seckill", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          order_num: "ORD-001",
          payment_id: "PAY-001",
          status: "pending",
        },
      }),
    })
  })

  await page.route("**/api/v1/orders/poll/ORD-001", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          status: "ready",
          order_num: "ORD-001",
          payment_id: "PAY-001",
          order: {
            order: {
              id: 9,
              order_num: "ORD-001",
              product_id: 1,
              status: 0,
              created_at: "2024-01-01T00:00:00Z",
            },
            payment: {
              payment_id: "PAY-001",
              status: "pending",
              amount_cents: 129900,
            },
          },
        },
      }),
    })
  })

  await page.route("**/api/v1/orders/9", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          order: {
            id: 9,
            order_num: "ORD-001",
            product_id: 1,
            status: 0,
            created_at: "2024-01-01T00:00:00Z",
          },
          payment: {
            payment_id: "PAY-001",
            status: "pending",
            amount_cents: 129900,
          },
        },
      }),
    })
  })

  await page.route("**/api/v1/coupons/mine**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          list: [],
          total: 0,
          page: 1,
          page_size: 100,
        },
      }),
    })
  })

  await page.goto("/login?redirect=/product/1")
  await page.getByPlaceholder("输入用户名").fill("alice")
  await page.getByPlaceholder("输入密码").fill("password-123")
  await page.getByRole("button", { name: "登录" }).click()
  await page.getByRole("button", { name: "立即抢购" }).click()

  await expect(page).toHaveURL(/\/orders\/9$/)
  await expect(page.getByRole("heading", { name: "订单详情" })).toBeVisible()
})
