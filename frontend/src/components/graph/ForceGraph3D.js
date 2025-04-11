// ForceGraph3D.js - Fixed version with improved node and edge handling

import React, { useEffect, useRef } from "react";
import ForceGraph3D from "3d-force-graph";

const CodebaseGraph3D = ({ graphData }) => {
  const containerRef = useRef(null);
  const graphRef = useRef(null);

  useEffect(() => {
    // Validate inputs
    if (!containerRef.current) {
      console.warn("Graph3D container not ready");
      return;
    }

    if (!graphData || !graphData.nodes || graphData.nodes.length === 0) {
      console.warn("Graph3D: No valid nodes found in graphData.");
      if (containerRef.current) {
        containerRef.current.innerHTML =
          '<div style="padding: 20px; color: grey; text-align: center;">No graph data to display</div>';
      }
      return;
    }

    // Clear any "no data" message if data is now present
    if (containerRef.current.innerHTML.includes("No graph data")) {
      containerRef.current.innerHTML = "";
    }

    // --- Data Transformation ---
    // Use a Map for faster lookup of nodes by ID
    const nodeMap = new Map();

    // Process nodes
    const nodes = graphData.nodes
      .filter((node) => node && node.id) // Filter out invalid nodes
      .map((node) => {
        // Determine node color based on type
        let color = "#aaaaaa"; // Default
        let nodeSize = 1; // Default size

        // Customize based on node type
        if (node.type === "file") {
          color = "#FFC107"; // Amber for files
          nodeSize = 1.5;
        } else if (node.type === "directory") {
          color = "#2196F3"; // Blue for directories
          nodeSize = 2;
        } else if (node.type === "component") {
          color = "#4CAF50"; // Green for components
          nodeSize = 1.8;
        }

        // Apply size from graphData if available
        if (node.size) {
          nodeSize = node.size / 10; // Scale down as 3D doesn't need as large values
        }

        const nodeData = {
          id: node.id,
          name: node.label || node.id.split("/").pop() || node.id, // Use label or extract filename
          description: node.description || `Type: ${node.type || "Unknown"}`,
          type: node.type || "unknown",
          color: node.color || color,
          val: nodeSize,
        };

        // Store in map for edge creation
        nodeMap.set(node.id, nodeData);
        return nodeData;
      });

    console.log(`ForceGraph3D: Prepared ${nodes.length} nodes`);

    // Process links/edges
    const links = [];

    // Add a link between nodes
    const addLink = (source, target, label, type) => {
      // Validate both nodes exist
      if (nodeMap.has(source) && nodeMap.has(target)) {
        // Use sourceNode and targetNode objects directly
        links.push({
          source,
          target,
          label: label || type || "related",
          type: type || "related",
          // Set color based on relationship type
          color:
            type === "IMPORTS" || type === "imports"
              ? "#FF5722"
              : type === "contains"
              ? "#9C27B0"
              : "#757575",
        });
      } else {
        console.debug(
          `Graph3D: Skipping link - Node not found: ${source} -> ${target}`
        );
      }
    };

    // Handle both edge formats: edges[] and relationships[]
    if (graphData.edges && Array.isArray(graphData.edges)) {
      graphData.edges.forEach((edge) => {
        if (edge.source && edge.target) {
          addLink(edge.source, edge.target, edge.label, edge.type);
        }
      });
    } else if (
      graphData.relationships &&
      Array.isArray(graphData.relationships)
    ) {
      graphData.relationships.forEach((rel) => {
        if (rel.source && rel.targets && Array.isArray(rel.targets)) {
          rel.targets.forEach((targetInfo) => {
            // Skip null targets
            if (targetInfo && targetInfo.target) {
              addLink(
                rel.source,
                targetInfo.target,
                targetInfo.label,
                targetInfo.type
              );
            }
          });
        }
      });
    }

    console.log(`ForceGraph3D: Prepared ${links.length} links`);

    // Debugging output
    if (links.length === 0) {
      console.warn(
        "No valid links found in graph data. This may result in an unconnected graph."
      );
    }

    // --- Initialize or Update 3D Force Graph ---
    try {
      // Create a new graph if not already created
      if (!graphRef.current) {
        graphRef.current = ForceGraph3D()(containerRef.current)
          .backgroundColor("#111111")
          .showNavInfo(true);
      }

      // Apply data and settings
      graphRef.current
        .graphData({ nodes, links })
        .nodeLabel(
          (
            node
          ) => `<div style="background: rgba(0,0,0,0.7); padding: 8px; border-radius: 4px; font-family: sans-serif;">
            <div style="font-weight: bold; margin-bottom: 4px; color: ${node.color}">${node.name}</div>
            <div style="font-size: 12px; color: #eee">${node.type}</div>
            <div style="font-size: 10px; margin-top: 2px; color: #aaa">${node.id}</div>
         </div>`
        )
        .linkLabel(
          (
            link
          ) => `<div style="background: rgba(0,0,0,0.7); padding: 4px; border-radius: 4px; font-family: sans-serif;">
            <div style="font-size: 10px; color: ${link.color}">${
            link.type || "related"
          }</div>
         </div>`
        )
        .nodeRelSize(5)
        .nodeColor((node) => node.color)
        .nodeVal((node) => node.val)
        .linkWidth(1.5)
        .linkColor((link) => link.color)
        .linkDirectionalParticles(1)
        .linkDirectionalParticleWidth((link) =>
          link.type === "IMPORTS" || link.type === "imports" ? 2 : 0
        )
        .linkDirectionalParticleSpeed(0.005)
        .linkDirectionalArrowLength(3.5)
        .linkDirectionalArrowRelPos(1)
        .width(containerRef.current.clientWidth)
        .height(containerRef.current.clientHeight)
        .onNodeClick((node) => {
          // Aim at node from outside
          const distance = 50;
          const distRatio = 1 + distance / Math.hypot(node.x, node.y, node.z);

          graphRef.current.cameraPosition(
            {
              x: node.x * distRatio,
              y: node.y * distRatio,
              z: node.z * distRatio,
            }, // new position
            node, // lookAt
            1000 // transition duration
          );
        });
    } catch (error) {
      console.error("Error initializing or updating ForceGraph3D:", error);
      if (containerRef.current) {
        containerRef.current.innerHTML = `<div style="padding: 20px; color: red; text-align: center;">Error rendering 3D graph: ${error.message}</div>`;
      }
    }

    // Handle resize
    const handleResize = () => {
      if (graphRef.current && containerRef.current) {
        graphRef.current
          .width(containerRef.current.clientWidth)
          .height(containerRef.current.clientHeight);
      }
    };

    window.addEventListener("resize", handleResize);

    // Cleanup
    return () => {
      window.removeEventListener("resize", handleResize);
      // No explicit destroy method for ForceGraph3D
      if (containerRef.current) {
        try {
          // Try to clear the canvas
          const canvas = containerRef.current.querySelector("canvas");
          if (canvas) {
            const ctx = canvas.getContext("2d");
            if (ctx) ctx.clearRect(0, 0, canvas.width, canvas.height);
          }
        } catch (e) {
          console.warn("Error during cleanup:", e);
        }
      }
      graphRef.current = null;
    };
  }, [graphData]); // Dependency array

  return (
    <div
      ref={containerRef}
      className="w-full bg-gray-900 rounded-lg border border-gray-700 overflow-hidden"
      style={{ height: "600px" }}
    />
  );
};

export default CodebaseGraph3D;
