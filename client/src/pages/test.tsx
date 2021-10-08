import { TouchBar } from "@/components/TouchBar"
import React from "react"

export default function TestPage() {
  return (
    <div
      style={{
        position: "absolute",
        top: "50%",
        left: "50%",
        width: "100%",
        padding: 20,
        transform: "translate(-50%, -50%)",
      }}
    >
      <TouchBar label="滑动发送" />
    </div>
  )
}
