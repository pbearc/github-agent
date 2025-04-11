// CodebaseGraph.js - Updated version with improved node and edge handling

import React, { useEffect, useRef } from "react";
import cytoscape from "cytoscape";
import cola from "cytoscape-cola";

cytoscape.use(cola);

const CodebaseGraph = ({ graphData }) => {
  const containerRef = useRef(null);
  const cyRef = useRef(null); // Ref to hold the Cytoscape instance

  useEffect(() => {
    // Ensure container exists
    if (!containerRef.current) {
      console.warn("Cytoscape container not ready.");
      return;
    }

    // If no data, potentially clear the graph or do nothing
    if (!graphData || !graphData.nodes || graphData.nodes.length === 0) {
      console.log("Cytoscape: No graph data provided.");
      // Destroy existing instance if graphData becomes null/empty
      if (cyRef.current) {
        console.log("Destroying Cytoscape instance due to empty data.");
        cyRef.current.destroy();
        cyRef.current = null;
      }
      // Clear the container visually if needed
      containerRef.current.innerHTML =
        '<div style="padding: 20px; color: grey;">No data to display</div>';
      return;
    }

    // Clear any "no data" message if data is now present
    if (containerRef.current.innerHTML.includes("No data to display")) {
      containerRef.current.innerHTML = "";
    }

    // --- Data Transformation ---
    const elements = [];
    const nodeIds = new Set();

    // Process nodes
    if (graphData.nodes && Array.isArray(graphData.nodes)) {
      graphData.nodes.forEach((node) => {
        if (node.id) {
          // Define colors based on node type
          let color = "#999999"; // Default color
          let shape = "ellipse"; // Default shape

          // Set color based on node type
          if (node.type === "file") {
            color = "#FFCA28"; // Yellow for files
            shape = "round-rectangle";
          } else if (node.type === "directory") {
            color = "#42A5F5"; // Blue for directories
            shape = "diamond";
          } else if (node.type === "component") {
            color = "#66BB6A"; // Green for components
          }

          elements.push({
            data: {
              id: node.id,
              label: node.label || node.id,
              type: node.type || "component",
              color: color,
              shape: shape,
              nodeSize: node.size || 30, // Default size if not provided
            },
            classes: node.type || "component",
          });

          nodeIds.add(node.id);
        }
      });
    }

    // Helper function to add edges
    const addEdge = (source, target, label, type) => {
      if (nodeIds.has(source) && nodeIds.has(target)) {
        // Set edge style based on type
        let lineStyle = "solid";
        let color = "#999999";

        if (type === "IMPORTS" || type === "imports") {
          color = "#FF5722"; // Orange for imports
        } else if (type === "contains") {
          color = "#7E57C2"; // Purple for contains relationship
          lineStyle = "dashed";
        }

        elements.push({
          data: {
            id: `${source}-${target}`,
            source: source,
            target: target,
            label: label || type || "",
            type: type || "relates",
            lineStyle: lineStyle,
            color: color,
          },
          classes: `edge ${type || "relates"}`,
        });
      }
    };

    // Process edges - handle both edges and relationships formats
    if (graphData.edges && Array.isArray(graphData.edges)) {
      graphData.edges.forEach((edge) => {
        if (edge.source && edge.target) {
          addEdge(edge.source, edge.target, edge.label, edge.type);
        }
      });
    } else if (
      graphData.relationships &&
      Array.isArray(graphData.relationships)
    ) {
      graphData.relationships.forEach((rel) => {
        if (rel.source && rel.targets && Array.isArray(rel.targets)) {
          rel.targets.forEach((targetInfo) => {
            // Only add if target is not null
            if (targetInfo && targetInfo.target) {
              addEdge(
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

    // Print debug info for developer
    console.log(`Cytoscape: Rendering ${elements.length} total elements`);
    console.log(
      `Nodes: ${nodeIds.size}, Edges: ${elements.length - nodeIds.size}`
    );

    // --- Initialize or Update Cytoscape ---
    try {
      // Destroy previous instance if it exists
      if (cyRef.current) {
        console.log("Destroying previous Cytoscape instance before update.");
        cyRef.current.destroy();
        cyRef.current = null;
      }

      // Create a new Cytoscape instance
      cyRef.current = cytoscape({
        container: containerRef.current,
        elements: elements,
        style: [
          // Node styles
          {
            selector: "node",
            style: {
              "label": "data(label)",
              "color": "#fff",
              "font-size": "12px",
              "font-weight": "bold",
              "text-valign": "center",
              "text-halign": "center",
              "background-color": "data(color)",
              "text-outline-color": "#222",
              "text-outline-width": "1px",
              "width": "data(nodeSize)",
              "height": "data(nodeSize)",
              "shape": "data(shape)",
              "text-wrap": "wrap",
              "text-max-width": "80px",
              "text-margin-y": "-5px",
              "text-background-opacity": 0.5,
              "text-background-color": "#000",
              "text-background-padding": "2px",
            },
          },

          // Node-type specific styles
          {
            selector: "node.file",
            style: {
              "shape": "round-rectangle",
            },
          },
          {
            selector: "node.directory",
            style: {
              "shape": "diamond",
            },
          },

          // Edge styles
          {
            selector: "edge",
            style: {
              "width": 2,
              "line-color": "data(color)",
              "target-arrow-color": "data(color)",
              "target-arrow-shape": "triangle",
              "curve-style": "bezier",
              "line-style": "data(lineStyle)",
              "text-background-opacity": 0.6,
              "text-background-color": "#000",
              "text-background-padding": "2px",
            },
          },

          // Specific edge types
          {
            selector: "edge.imports",
            style: {
              "line-style": "solid",
              "label": "imports",
              "text-rotation": "autorotate",
              "font-size": "8px",
              "color": "#fff",
              "text-outline-color": "#000",
              "text-outline-width": "1px",
            },
          },
          {
            selector: "edge.contains",
            style: {
              "line-style": "dashed",
              "label": "contains",
              "text-rotation": "autorotate",
              "font-size": "8px",
              "color": "#fff",
            },
          },

          // Hover states
          {
            selector: "node:selected",
            style: {
              "border-width": 3,
              "border-color": "#d33",
              "border-opacity": 0.8,
              "background-color": "data(color)",
              "font-size": "14px",
            },
          },
          {
            selector: "edge:selected",
            style: {
              "width": 4,
              "line-color": "#d33",
              "target-arrow-color": "#d33",
            },
          },
        ],
        // Set a default layout - actual layout will be run later
        layout: { name: "preset" },
        // Interaction settings
        wheelSensitivity: 0.3,
        minZoom: 0.2,
        maxZoom: 3,
        userZoomingEnabled: true,
        userPanningEnabled: true,
      });

      // --- Event Handlers ---
      cyRef.current.on("tap", "node", function (evt) {
        if (!cyRef.current) return;
        const node = evt.target;
        console.log(
          `Node Tapped: ${node.id()} - Label: ${node.data(
            "label"
          )}, Type: ${node.data("type")}`
        );
      });

      cyRef.current.on("tap", "edge", function (evt) {
        if (!cyRef.current) return;
        const edge = evt.target;
        console.log(
          `Edge Tapped: ${edge.id()} - Type: ${edge.data(
            "type"
          )}, Source: ${edge.data("source")}, Target: ${edge.data("target")}`
        );
      });

      cyRef.current.on("mouseover", "node", function (evt) {
        if (!cyRef.current) return;
        const node = evt.target;
        node.style({
          "border-width": 3,
          "border-color": "#ff0",
        });
      });

      cyRef.current.on("mouseout", "node", function (evt) {
        if (!cyRef.current) return;
        const node = evt.target;
        node.style({
          "border-width": 1,
          "border-color": node.data("color"),
        });
      });

      // --- Run Layout ---
      console.log("Running Cola layout...");
      const layout = cyRef.current.layout({
        name: "cola",
        nodeSpacing: 120,
        edgeLengthVal: 180,
        animate: "end",
        animationDuration: 800,
        randomize: true,
        maxSimulationTime: 4000,
        fit: true,
        padding: 40,
        infinite: false,
      });

      layout.on("layoutstop", () => {
        console.log("Cola layout stopped.");
        if (cyRef.current) {
          cyRef.current.fit(undefined, 40);
          cyRef.current.center();
        }
      });

      layout.run();
    } catch (error) {
      console.error("Error initializing or running Cytoscape:", error);
      if (containerRef.current) {
        containerRef.current.innerHTML = `<div style="padding: 20px; color: red;">Error rendering graph: ${error.message}</div>`;
      }
    }

    // --- Cleanup ---
    return () => {
      if (cyRef.current) {
        cyRef.current.destroy();
        cyRef.current = null;
      }
    };
  }, [graphData]); // Dependency array

  // Control functions
  const runLayout = () => {
    if (cyRef.current) {
      cyRef.current
        .layout({
          name: "cola",
          nodeSpacing: 120,
          edgeLengthVal: 180,
          animate: true,
          randomize: true,
          maxSimulationTime: 3000,
          fit: true,
          padding: 40,
        })
        .run();
    }
  };

  const handleFit = () => {
    if (cyRef.current) {
      cyRef.current.fit(undefined, 40);
      cyRef.current.center();
    }
  };

  const handleZoomIn = () => {
    if (cyRef.current) {
      cyRef.current.zoom(cyRef.current.zoom() * 1.3);
      cyRef.current.center();
    }
  };

  const handleZoomOut = () => {
    if (cyRef.current) {
      cyRef.current.zoom(cyRef.current.zoom() / 1.3);
      cyRef.current.center();
    }
  };

  return (
    <div className="relative">
      <div
        ref={containerRef}
        className="w-full bg-gray-900 rounded-lg border border-gray-700"
        style={{ height: "600px" }}
      />

      {/* Controls */}
      <div className="absolute bottom-4 right-4 flex space-x-2">
        <button
          title="Re-run Layout"
          className="bg-dark-300 hover:bg-dark-200 text-gray-300 p-2 rounded-full"
          onClick={runLayout}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z"
              clipRule="evenodd"
            />
          </svg>
        </button>
        <button
          title="Fit Graph"
          className="bg-dark-300 hover:bg-dark-200 text-gray-300 p-2 rounded-full"
          onClick={handleFit}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M3 4a1 1 0 011-1h4a1 1 0 010 2H6.414l2.293 2.293a1 1 0 01-1.414 1.414L5 6.414V8a1 1 0 01-2 0V4zm9 1a1 1 0 110-2h4a1 1 0 011 1v4a1 1 0 01-2 0V6.414l-2.293 2.293a1 1 0 11-1.414-1.414L13.586 5H12zm-9 7a1 1 0 112 0v1.586l2.293-2.293a1 1 0 011.414 1.414L6.414 15H8a1 1 0 110 2H4a1 1 0 01-1-1v-4zm13-1a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 110-2h1.586l-2.293-2.293a1 1 0 111.414-1.414L15 13.586V12a1 1 0 011-1z"
              clipRule="evenodd"
            />
          </svg>
        </button>
        <button
          title="Zoom In"
          className="bg-dark-300 hover:bg-dark-200 text-gray-300 p-2 rounded-full"
          onClick={handleZoomIn}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z"
              clipRule="evenodd"
            />
          </svg>
        </button>
        <button
          title="Zoom Out"
          className="bg-dark-300 hover:bg-dark-200 text-gray-300 p-2 rounded-full"
          onClick={handleZoomOut}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M5 10a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1z"
              clipRule="evenodd"
            />
          </svg>
        </button>
      </div>
    </div>
  );
};

export default CodebaseGraph;
