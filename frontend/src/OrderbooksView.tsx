import React, { useEffect, useState, useRef } from "react";
import axios from "axios";

export default function OrderBooksView() {
  const [orderBooks, setOrderBooks] = useState<Record<string, any>>({});
  const [loading, setLoading] = useState(true);
  const [buySymbol, setBuySymbol] = useState("BTC");
  const [buyPrice, setBuyPrice] = useState("1");
  const [buyQty, setBuyQty] = useState("1");

  const wsRef = useRef<WebSocket | null>(null);
  const wsRetryTimer = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    fetchOrderBooks();
    connectWebSocket();

    return () => {
      wsRef.current?.close();
      if (wsRetryTimer.current) clearTimeout(wsRetryTimer.current);
    };
  }, []);

  const fetchOrderBooks = async () => {
    try {
      const res = await axios.get(`http://${process.env.REACT_APP_API_URL}/orderbook`);
      setOrderBooks(res.data);
      setLoading(false);
    } catch (err) {
      console.error("Failed to fetch orderbooks:", err);
      // 1 saniye sonra tekrar dene
      setTimeout(fetchOrderBooks, 1000);
      setLoading(true);
    }
  };

  const connectWebSocket = () => {
    const ws = new WebSocket(`ws://${process.env.REACT_APP_API_URL}/event`);
    wsRef.current = ws;

    ws.onopen = () => {
      setLoading(false);
      console.log("WebSocket connected");
      if (wsRetryTimer.current) {
        clearTimeout(wsRetryTimer.current);
        wsRetryTimer.current = null;
      }
    };

    ws.onclose = () => {
      setLoading(true);
      console.log("WebSocket closed, retrying in 1s...");
      wsRetryTimer.current = setTimeout(connectWebSocket, 1000);
    };

    ws.onerror = (err) => {
      console.error("WebSocket error", err);
      ws.close();
      setLoading(true);
    };

    ws.onmessage = (msg) => {
      try {
        const data = JSON.parse(msg.data);
        handleEvent(data);
      } catch (err) {
        console.error("Failed to parse WS message:", err);
      }
    };
  };

  const EPS = 1e-8;
  const priceEquals = (a: number, b: number) => Math.abs(a - b) < EPS;

  const handleEvent = (event: any) => {
    const { type, payload } = event;
    const { symbol, side, price, remaining, quantity } = payload;

    setOrderBooks((prev) => {
      const prevSymbol = prev[symbol] || {};
      const bidsArr = Array.isArray(prevSymbol.bids) ? [...prevSymbol.bids] : [];
      const asksArr = Array.isArray(prevSymbol.asks) ? [...prevSymbol.asks] : [];

      let isBid: boolean | null = null;
      if (side === "buy") isBid = true;
      else if (side === "sell") isBid = false;
      else if (type === "order_matched") {
        if (bidsArr.some((o: any) => priceEquals(o.price, price))) isBid = true;
        else if (asksArr.some((o: any) => priceEquals(o.price, price))) isBid = false;
        else return prev;
      } else {
        isBid = true;
      }

      let updatedSide = isBid ? bidsArr : asksArr;
      const findIdx = (arr: any[], p: number) => arr.findIndex((o) => priceEquals(o.price, p));

      if (type === "order_added") {
        const addQty = Number(remaining ?? quantity ?? 0);
        if (addQty <= 0) {
          const idx = findIdx(updatedSide, price);
          if (idx !== -1) updatedSide.splice(idx, 1);
        } else {
          const idx = findIdx(updatedSide, price);
          if (idx !== -1) {
            updatedSide[idx] = { ...updatedSide[idx], qty: (Number(updatedSide[idx].qty) || 0) + addQty };
          } else {
            updatedSide.push({ price: Number(price), qty: addQty });
          }
        }
      } else if (type === "order_matched") {
        const matchedQty = Number(quantity ?? remaining ?? 0);
        if (matchedQty > 0) {
          const idx = findIdx(updatedSide, price);
          if (idx !== -1) {
            const existing = updatedSide[idx];
            const newQty = (Number(existing.qty) || 0) - matchedQty;
            if (newQty > EPS) updatedSide[idx] = { ...existing, qty: newQty };
            else updatedSide.splice(idx, 1);
          }
        }
      } else {
        return prev;
      }

      updatedSide.sort((a, b) => (isBid ? b.price - a.price : a.price - b.price));

      const newBook = {
        bids: isBid ? updatedSide : bidsArr,
        asks: isBid ? asksArr : updatedSide,
      };

      return {
        ...prev,
        [symbol]: newBook,
      };
    });
  };

  const placeOrder = async (type: "buy" | "sell") => {
    if (!buyPrice || !buyQty) return alert("Price ve Quantity girin");

    try {
      await axios.post(`http://${process.env.REACT_APP_API_URL}/orders`, {
        symbol: buySymbol,
        side: type,
        price: parseFloat(buyPrice),
        quantity: parseFloat(buyQty),
      });
      alert(`${type === "buy" ? "Buy" : "Sell"} order sent!`);
    } catch (err) {
      console.error(`Failed to place ${type} order:`, err);
      alert(`Failed to place ${type} order`);
    }
  };

  if (loading) return <div>Loading order books...</div>;

  const containerStyle: React.CSSProperties = { padding: "24px" };
  const cardStyle: React.CSSProperties = {
    border: "1px solid #ccc",
    borderRadius: "16px",
    boxShadow: "0 2px 6px rgba(0,0,0,0.1)",
    padding: "16px",
    marginBottom: "24px",
    backgroundColor: "#fff",
  };
  const tableStyle: React.CSSProperties = { width: "100%", borderCollapse: "collapse", marginBottom: "12px" };
  const thTdStyle: React.CSSProperties = { border: "1px solid #ccc", padding: "4px 8px" };
  const gridStyle: React.CSSProperties = { display: "flex", gap: "16px" };
  const columnStyle: React.CSSProperties = { flex: 1 };
  const gridContainerStyle: React.CSSProperties = {
    maxWidth: "400px", display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "24px",
  };
  const formStyle: React.CSSProperties = { marginBottom: "24px", display: "flex", gap: "8px", alignItems: "center" };
  const inputStyle: React.CSSProperties = { padding: "4px 8px", border: "1px solid #ccc", borderRadius: "4px" };
  const buttonBuyStyle: React.CSSProperties = { padding: "6px 12px", borderRadius: "4px", cursor: "pointer", backgroundColor: "green", color: "#fff", border: "none" };
  const buttonSellStyle: React.CSSProperties = { padding: "6px 12px", borderRadius: "4px", cursor: "pointer", backgroundColor: "red", color: "#fff", border: "none" };
  const h1Style: React.CSSProperties = { fontWeight: "bold", marginBottom: "30px", textAlign: "center" };
  const h2Style: React.CSSProperties = { fontSize: "20px", fontWeight: "600", marginBottom: "8px" };
  const h3RedStyle: React.CSSProperties = { color: "red", fontWeight: "500", marginBottom: "4px" };
  const h3GreenStyle: React.CSSProperties = { color: "green", fontWeight: "500", marginBottom: "4px" };
  return (
    <div style={containerStyle}>
      <h1 style={h1Style}>Order Books</h1>

      <div style={formStyle}>
        <span>Symbol:</span>
        <input style={inputStyle} value={buySymbol} onChange={(e) => setBuySymbol(e.target.value)} />
        <span>Price:</span>
        <input style={inputStyle} defaultValue={1} value={buyPrice} onChange={(e) => setBuyPrice(e.target.value)} />
        <span>Qty:</span>
        <input style={inputStyle} defaultValue={1} value={buyQty} onChange={(e) => setBuyQty(e.target.value)} />
        <button style={buttonBuyStyle} onClick={() => placeOrder("buy")}>
          Buy
        </button>
        <button style={buttonSellStyle} onClick={() => placeOrder("sell")}>
          Sell
        </button>
      </div>

      <div style={gridContainerStyle}>
        {Object.entries(orderBooks).map(([symbol, book]) => (
          <div key={symbol} style={cardStyle}>
            <h2 style={h2Style}>{symbol}</h2>

            <div style={gridStyle}>
              <div style={columnStyle}>
                <h3 style={h3GreenStyle}>Bids</h3>
                <table style={tableStyle}>
                  <thead>
                    <tr>
                      <th style={thTdStyle}>Price</th>
                      <th style={thTdStyle}>Qty</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(book.bids || []).map((b: any, i: number) => (
                      <tr key={i} style={{ color: "green" }}>
                        <td style={thTdStyle}>{b.price}</td>
                        <td style={thTdStyle}>{b.qty}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div style={columnStyle}>
                <h3 style={h3RedStyle}>Asks</h3>
                <table style={tableStyle}>
                  <thead>
                    <tr>
                      <th style={thTdStyle}>Price</th>
                      <th style={thTdStyle}>Qty</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(book.asks || []).map((a: any, i: number) => (
                      <tr key={i} style={{ color: "red" }}>
                        <td style={thTdStyle}>{a.price}</td>
                        <td style={thTdStyle}>{a.qty}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}