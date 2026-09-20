import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { OENClient, RPCError } from '../services/rpcClient'
import type { ChainInfo, BlockInfo, FeeEstimate } from '../types/rpc'

describe('OENClient JSON-RPC 2.0 Client (TDD)', () => {
  let client: OENClient
  const mockEndpoint = 'http://127.0.0.1:8545'

  beforeEach(() => {
    client = new OENClient(mockEndpoint)
    vi.restoreAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('initializes with custom and default endpoints', () => {
    expect(client.getEndpoint()).toBe(mockEndpoint)
    client.setEndpoint('http://node.oenexa.org:8545')
    expect(client.getEndpoint()).toBe('http://node.oenexa.org:8545')

    const defaultClient = new OENClient()
    expect(defaultClient.getEndpoint()).toBe('http://localhost:8545')
  })

  it('calls oen_chainInfo and parses response', async () => {
    const mockInfo: ChainInfo = {
      chainId: 'oenexa-mainnet',
      height: 105,
      tipHash: 'a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90',
      validators: 4,
      round: 1,
      mempoolSize: 2,
      baseFee: 1000000,
      gasLimit: 30000000,
      gasTargetRatio: 0.5,
    }

    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jsonrpc: '2.0',
        id: 1,
        result: mockInfo,
      }),
    })
    globalThis.fetch = mockFetch as any

    const info = await client.getChainInfo()
    expect(info).toEqual(mockInfo)
    expect(mockFetch).toHaveBeenCalledWith(mockEndpoint, expect.objectContaining({
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: expect.stringContaining('"method":"oen_chainInfo"'),
    }))
  })

  it('calls oen_blockByHeight and parses block details', async () => {
    const mockBlock: BlockInfo = {
      height: 42,
      hash: '0000abcd12345678901234567890123456789012345678901234567890abcdef',
      parentHash: '0000000000000000000000000000000000000000000000000000000000000000',
      timestamp: 1725890000,
      proposer: 'OEN1A2B3C4D5E6F7890123456789012345678901',
      txCount: 1,
      transactions: [
        {
          hash: 'txhash1234567890abcdef',
          type: 'transfer',
          from: 'OEN1A2B3C4D5E6F7890123456789012345678901',
          to: 'OEN9Z8Y7X6W5V4U3210987654321098765432109',
          value: '1000000000000000000',
          nonce: 0,
          gasLimit: 21000,
          maxFeePerGas: 1000000,
        },
      ],
      gasUsed: 21000,
      baseFee: 1000000,
    }

    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jsonrpc: '2.0',
        id: 2,
        result: mockBlock,
      }),
    })
    globalThis.fetch = mockFetch as any

    const block = await client.getBlockByHeight(42)
    expect(block).toEqual(mockBlock)
    expect(mockFetch).toHaveBeenCalledWith(mockEndpoint, expect.objectContaining({
      body: expect.stringContaining('"method":"oen_blockByHeight"'),
    }))
  })

  it('calls oen_blockByHash correctly', async () => {
    const mockHash = '0000abcd12345678901234567890123456789012345678901234567890abcdef'
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jsonrpc: '2.0',
        id: 3,
        result: { height: 10, hash: mockHash },
      }),
    })
    globalThis.fetch = mockFetch as any

    const res = await client.getBlockByHash(mockHash)
    expect(res.height).toBe(10)
  })

  it('calls oen_getBalance and returns balance string', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jsonrpc: '2.0',
        id: 4,
        result: '50000000000000000000',
      }),
    })
    globalThis.fetch = mockFetch as any

    const bal = await client.getBalance('OEN123')
    expect(bal).toBe('50000000000000000000')
  })

  it('calls oen_getTransactionCount and returns nonce', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jsonrpc: '2.0',
        id: 5,
        result: 7,
      }),
    })
    globalThis.fetch = mockFetch as any

    const nonce = await client.getTransactionCount('OEN123')
    expect(nonce).toBe(7)
  })

  it('calls oen_sendRawTransaction and returns tx hash', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jsonrpc: '2.0',
        id: 6,
        result: '0xhash999',
      }),
    })
    globalThis.fetch = mockFetch as any

    const txHash = await client.sendRawTransaction('deadbeef')
    expect(txHash).toBe('0xhash999')
  })

  it('calls oen_feeEstimate and parses FeeEstimate', async () => {
    const mockFee: FeeEstimate = {
      baseFee: 1000000,
      slow: 1000000,
      standard: 1200000,
      fast: 1500000,
    }

    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jsonrpc: '2.0',
        id: 7,
        result: mockFee,
      }),
    })
    globalThis.fetch = mockFetch as any

    const estimate = await client.getFeeEstimate()
    expect(estimate).toEqual(mockFee)
  })

  it('calls oen_gasPrice and returns gas price number', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jsonrpc: '2.0',
        id: 8,
        result: 1050000,
      }),
    })
    globalThis.fetch = mockFetch as any

    const gp = await client.getGasPrice()
    expect(gp).toBe(1050000)
  })

  it('throws RPCError when server returns JSON-RPC error', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jsonrpc: '2.0',
        id: 9,
        error: {
          code: -32601,
          message: 'Method not found',
        },
      }),
    })
    globalThis.fetch = mockFetch as any

    await expect(client.getChainInfo()).rejects.toThrow(RPCError)
    await expect(client.getChainInfo()).rejects.toThrow('Method not found')
  })

  it('throws network error when fetch fails', async () => {
    const mockFetch = vi.fn().mockRejectedValue(new Error('Connection refused'))
    globalThis.fetch = mockFetch as any

    await expect(client.getChainInfo()).rejects.toThrow('Connection refused')
  })

  it('dispatches batch RPC calls successfully', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => [
        { jsonrpc: '2.0', id: 1, result: { height: 10 } },
        { jsonrpc: '2.0', id: 2, result: '100000' },
      ],
    })
    globalThis.fetch = mockFetch as any

    const batch = await client.dispatchBatch([
      { method: 'oen_chainInfo', params: [] },
      { method: 'oen_getBalance', params: ['OEN123'] },
    ])

    expect(batch).toHaveLength(2)
    expect(batch[0]).toEqual({ height: 10 })
    expect(batch[1]).toBe('100000')
  })
})
