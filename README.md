# Price Ticker

For each minute calculates index price based on most relevant prices from price streams.

### Solution

`/internal/task/price_stream.go`

`/cmd/index_price/main.go`

# Algorithm

#### Input

`TickerPrice` streams

#### Output

timestamp, index price

#### Description

- for each solid minute timestamp `t` in price stream can be many `TickerPrices`
  with `TickerPrice.Time <= t`
- from each stream we take most relevant `TickerPrice` (i.e., `TickerPrice.Time <= t` and `TickerPrice.Time` is the
  nearest timestamp to `t`)
    - if `t - TickerPrice.Time <= relevantOffset` we use this `TickerPrice`
    - if `t - TickerPrice.Time  > relevantOffset` we skip this `TickerPrice` from index price calculation
    - where `relevantOffset` is parameter when `TickerPrice` with `TickerPrice.Time` in window `[t - relevantOffset, t]`
      is considered valid for current index price calculation
- if `TickerPrice.Time > t` we don't use such `TickerPrice` in index price calculation for timestamp `t`
- if stream is closed task description doesn't provide advice for this situation
    - we can use the last `TickerPrice` some period of time in range of `[t - relevantOffset, t]`
    - we can delete this stream from streams list immediately

#### Example

```

stream 1:    |  A0    B0  B1 B2|   B3 C0   <- TickerPrice is already in prices channel in point of timepline
             |        |        |
             |        |        |
stream 2:    | A0 A1  | A2  B0 | C0
             |        |        | 
             |        |        |
stream 3:    |A0    (closed)   | 
             |        |        |
       [relevantOffset]        |    <- B relevant offset
             |  [relevantOffset]    <- C relevant offset
             |        |        | 
---------------------------------------> timeline (seconds) 
             A        B        C 
             0        60      120 
```

`A`, `B`, `C` solid minute timestamps

`A[N]`, `B[N]`, `C[N]` is `TickerPrice`

- which `TickerPrice.Time >= A` and `< B` for `A[N]`
- which `TickerPrice.Time >= B` and `< C` for `B[N]`
- which `TickerPrice.Time >= C` for `C[N]`

For timestamp `B` will be chosen prices:

- stream1: `B0 (B0 == B, skipping A0)`
- stream2: `A1 (A1 <= B, skipping A0, A2 isn't present in stream channel yet)`
- stream3: `A0 (A0 < B and A0 in window [B - relevantOffset:B] yet)`

For timestamp `C` will be chosen prices:

- stream1: `B2 (B2 < C, skipping B0, B1)`
- stream2: `B0 (B0 < C, C0 isn't present in stream channel yet)`
- stream3: `ignored because A0 not in window [C - relevantOffset:C]`

# Run

`./run.sh`

#### Output example

```
1649671740 99.56
1649671800 99.94
1649671860 100.67
```
