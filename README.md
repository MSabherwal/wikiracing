# wikiracing
Fastest pathfinder between two wikipedia pages

## Design
Pathfinding is done through Bi-directional BFS. Each approach,
(forward and backwards) Queries the wikipedia API for links on
the page or titles that link to the page.

Using these, a feeback loop system is created where links have
child links, and those children have links, etc..

A map of node to parent is created on both sides of the search,
where the key is the current node and the value is the parent.

When the final value  OR a node that exists in the opposite
search's map is found,a full path can be generated and searching 
is stopped.
