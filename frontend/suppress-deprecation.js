// Suppress specific deprecation warnings
const originalEmit = process.emit;

process.emit = function(name, data, ...args) {
  // Suppress DEP0060: util._extend is deprecated
  if (
    name === 'warning' && 
    typeof data === 'object' && 
    data.name === 'DeprecationWarning' &&
    data.code === 'DEP0060'
  ) {
    return false;
  }
  
  return originalEmit.call(process, name, data, ...args);
};