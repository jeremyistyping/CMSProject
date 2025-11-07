# 🧪 Rupiah Input Testing Guide

## 🎯 **Test Cases for Amount Input**

### **1. Basic Input Tests**

#### **Test Case 1: Simple Numbers**
- **Input**: `1000`
- **Expected Display**: `Rp 1.000`
- **Status**: ✅ Should format correctly

#### **Test Case 2: Large Numbers**  
- **Input**: `1500000`
- **Expected Display**: `Rp 1.500.000`
- **Status**: ✅ Should format correctly

#### **Test Case 3: Very Large Numbers**
- **Input**: `15000000000`
- **Expected Display**: `Rp 15.000.000.000`
- **Status**: ✅ Should format correctly

### **2. Edge Cases**

#### **Test Case 4: Zero**
- **Input**: `0`
- **Expected Display**: Empty field or `Rp 0`
- **Status**: ✅ Should handle gracefully

#### **Test Case 5: Negative Numbers** 
- **Input**: `-1000`
- **Expected**: Should be prevented (min=0)
- **Status**: ✅ Should not allow negative

#### **Test Case 6: Decimal Input**
- **Input**: `1000.50`
- **Expected**: `Rp 1.000` (decimals removed)
- **Status**: ✅ Should truncate decimals

### **3. Copy/Paste Tests**

#### **Test Case 7: Paste Formatted Number**
- **Input**: Paste `Rp 1.500.000`
- **Expected**: Should parse and display correctly
- **Status**: ✅ Should parse existing format

#### **Test Case 8: Paste Plain Number**
- **Input**: Paste `1500000`  
- **Expected**: `Rp 1.500.000`
- **Status**: ✅ Should format pasted numbers

#### **Test Case 9: Paste Invalid Format**
- **Input**: Paste `abc123`
- **Expected**: Should ignore or revert
- **Status**: ✅ Should handle invalid input

### **4. User Interaction Tests**

#### **Test Case 10: Stepper Buttons**
- **Action**: Click increment/decrement
- **Expected**: Should format automatically
- **Status**: ✅ Should work with steppers

#### **Test Case 11: Backspace/Delete**
- **Action**: Delete digits from formatted number
- **Expected**: Should reformat remaining digits
- **Status**: ✅ Should reformat after deletion

#### **Test Case 12: Mobile Input**
- **Device**: Mobile/tablet
- **Expected**: Should display numeric keypad
- **Status**: ✅ Should be mobile-friendly

## 🔧 **Manual Testing Steps**

### **Step 1: Basic Functionality**
1. Open deposit form
2. Click on Amount field
3. Type: `1000000`
4. **Verify**: Shows `Rp 1.000.000`
5. **Result**: ✅ Pass / ❌ Fail

### **Step 2: Parse Functionality** 
1. Clear the field
2. Paste: `Rp 2.500.000`
3. **Verify**: Field accepts and shows correctly
4. **Result**: ✅ Pass / ❌ Fail

### **Step 3: Form Submission**
1. Enter amount: `1000000`
2. Fill other fields
3. Click "Process Deposit"
4. **Verify**: Correct amount sent to backend
5. **Result**: ✅ Pass / ❌ Fail

### **Step 4: Preview Display**
1. Enter amount: `5000000`
2. **Verify**: Double-entry preview shows `+Rp 5.000.000`
3. **Verify**: Balance preview shows correct new balance
4. **Result**: ✅ Pass / ❌ Fail

## 🐛 **Common Issues to Check**

### **Issue 1: Field Too Narrow**
- **Problem**: Amount text gets cut off
- **Solution**: ✅ Fixed with `w="100%"` and proper padding
- **Test**: Enter `15000000000` and verify full display

### **Issue 2: Parse Errors**
- **Problem**: Format/parse cycle creates errors
- **Solution**: ✅ Improved parse function
- **Test**: Type and reformat multiple times

### **Issue 3: Indonesian Format**
- **Problem**: Uses comma instead of dot separators
- **Solution**: ✅ Using `toLocaleString('id-ID')`
- **Test**: Verify dots, not commas as thousand separators

### **Issue 4: Modal Size**
- **Problem**: Modal too small for content
- **Solution**: ✅ Changed from `xl` to `2xl`
- **Test**: Check all content fits without scrolling

## 📱 **Cross-Browser Testing**

### **Chrome**: ✅ / ❌
### **Firefox**: ✅ / ❌  
### **Safari**: ✅ / ❌
### **Edge**: ✅ / ❌
### **Mobile Chrome**: ✅ / ❌
### **Mobile Safari**: ✅ / ❌

## 🎯 **Performance Testing**

### **Large Numbers**
- **Test**: Enter `999999999999`
- **Expected**: Should handle without lag
- **Result**: ✅ / ❌

### **Rapid Typing**
- **Test**: Type very fast
- **Expected**: Should format smoothly
- **Result**: ✅ / ❌

## 📋 **Checklist for Release**

- [ ] All basic input tests pass
- [ ] Edge cases handled properly  
- [ ] Copy/paste works correctly
- [ ] Mobile-friendly
- [ ] Cross-browser compatible
- [ ] Performance acceptable
- [ ] UI looks professional
- [ ] Help text is clear
- [ ] Error handling works
- [ ] Integration with form submission works
