import { useEffect, useRef, memo } from 'react'
import { useLanguage } from '../contexts/LanguageContext'

interface TradingViewScreenerProps {
  height?: number | string
  width?: number | string
}

function TradingViewScreenerComponent({
  height = 600,
  width = '100%',
}: TradingViewScreenerProps) {
  const { language } = useLanguage()
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!containerRef.current) return

    // Clear container
    containerRef.current.innerHTML = ''

    // Create widget container
    const widgetContainer = document.createElement('div')
    widgetContainer.className = 'tradingview-widget-container'
    widgetContainer.style.height = '100%'
    widgetContainer.style.width = '100%'

    const widgetDiv = document.createElement('div')
    widgetDiv.className = 'tradingview-widget-container__widget'
    widgetContainer.appendChild(widgetDiv)
    containerRef.current.appendChild(widgetContainer)

    // Load TradingView Script
    const script = document.createElement('script')
    script.src =
      'https://s3.tradingview.com/external-embedding/embed-widget-screener.js'
    script.type = 'text/javascript'
    script.async = true
    script.innerHTML = JSON.stringify({
      width: '100%',
      height: '100%',
      defaultColumn: 'overview',
      screener_type: 'crypto_mkt',
      displayCurrency: 'USD',
      colorTheme: 'dark',
      locale: language === 'zh' ? 'zh_CN' : 'en',
      isTransparent: false,
    })

    widgetContainer.appendChild(script)

    return () => {
      if (containerRef.current) {
        containerRef.current.innerHTML = ''
      }
    }
  }, [language])

  return (
    <div
      ref={containerRef}
      className="tradingview-widget-container"
      style={{ height: typeof height === 'number' ? height : height, width }}
    />
  )
}

export const TradingViewScreener = memo(TradingViewScreenerComponent)
