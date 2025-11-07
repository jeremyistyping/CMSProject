# Payment Components with Journal Integration

This directory contains React components that provide a comprehensive payment management interface integrated with the SSOT (Single Source of Truth) journal system.

## Components Overview

### 1. PaymentDashboard
**File**: `PaymentDashboard.tsx`

The main dashboard component that provides both analytics and payment management functionality.

**Features**:
- **Analytics Dashboard Tab**: Payment analytics, trends, and performance metrics
- **Payment Management Tab**: Journal integration metrics and payment creation
- **View Modes**: Dashboard, Create Payment, Payment Details
- **Integration Metrics**: Shows payment-to-journal integration statistics
- **Quick Actions**: Create payments with journal entries

**Usage**:
```typescript
import { PaymentDashboard } from '@/components/payments';

<PaymentDashboard />
```

### 2. PaymentWithJournalForm
**File**: `PaymentWithJournalForm.tsx`

A comprehensive form for creating payments with automatic journal entry creation.

**Features**:
- **Real-time Journal Preview**: Shows journal entries as you type
- **Auto-preview**: Automatically generates preview when form changes
- **Journal Options**: Configure journal creation settings
- **Payment Types**: Support for RECEIVE and SEND payments
- **Multiple Payment Methods**: Bank transfer, cash, cards, digital wallets
- **Related Entity Links**: Customer, vendor, and invoice associations

**Props**:
```typescript
interface PaymentWithJournalFormProps {
  onSubmit?: (paymentId: string) => void;
  onCancel?: () => void;
  initialData?: Partial<CreatePaymentWithJournalRequest>;
}
```

**Usage**:
```typescript
import { PaymentWithJournalForm } from '@/components/payments';

<PaymentWithJournalForm
  onSubmit={(paymentId) => console.log('Payment created:', paymentId)}
  onCancel={() => console.log('Form cancelled')}
/>
```

### 3. PaymentDetailsView
**File**: `PaymentDetailsView.tsx`

Detailed view of a payment with its associated journal entries and account balance changes.

**Features**:
- **Tabbed Interface**: Payment details, journal entry, account balances
- **Payment Information**: Complete payment details with status
- **Journal Entry Details**: Associated journal entry information
- **Account Balance Changes**: Shows how the payment affected account balances
- **Payment Actions**: Edit, reverse payments
- **Real-time Balance Refresh**: Update account balances on demand

**Props**:
```typescript
interface PaymentDetailsViewProps {
  paymentId: string;
  onEdit?: (paymentId: string) => void;
  onReverse?: (paymentId: string) => void;
}
```

**Usage**:
```typescript
import { PaymentDetailsView } from '@/components/payments';

<PaymentDetailsView
  paymentId="payment-123"
  onEdit={(id) => handleEdit(id)}
  onReverse={(id) => handleReverse(id)}
/>
```

### 4. JournalEntryPreview
**File**: `JournalEntryPreview.tsx`

Reusable component for displaying journal entry information with account balance updates.

**Features**:
- **Journal Entry Display**: Shows entry number, status, and balance information
- **Journal Lines Table**: Detailed debit/credit breakdown
- **Account Balance Updates**: Before/after balance changes
- **Expandable/Collapsible**: Can be toggled for space saving
- **Preview Mode**: Special styling for preview vs actual entries

**Props**:
```typescript
interface JournalEntryPreviewProps {
  journalResult: PaymentJournalResult | null;
  title?: string;
  isPreview?: boolean;
  showAccountUpdates?: boolean;
  isExpanded?: boolean;
  onToggleExpand?: () => void;
}
```

**Usage**:
```typescript
import { JournalEntryPreview } from '@/components/payments';

<JournalEntryPreview
  journalResult={journalData}
  title="Journal Entry Preview"
  isPreview={true}
  showAccountUpdates={true}
/>
```

## Integration with Backend Services

These components integrate with the backend through the `paymentService`:

### Required Service Methods
- `createPaymentWithJournal()` - Create payment with journal entry
- `previewJournalEntry()` - Generate journal preview
- `getPaymentWithJournal()` - Get payment with journal info
- `reversePayment()` - Reverse payment and journal entries
- `getAccountBalanceUpdates()` - Get balance changes
- `getIntegrationMetrics()` - Get integration statistics
- `refreshAccountBalances()` - Refresh account balances

### Data Types
The components use these TypeScript interfaces:
- `PaymentJournalResult` - Result of journal operations
- `PaymentWithJournalInfo` - Payment with journal information
- `JournalEntry` - Journal entry details
- `AccountBalanceUpdate` - Account balance change information
- `PaymentIntegrationMetrics` - Integration statistics

## Styling and Theming

All components use Chakra UI for styling and support:
- **Light/Dark Mode**: Automatic color mode switching
- **Responsive Design**: Mobile-friendly layouts
- **Consistent Theming**: Follows Chakra UI color schemes
- **Icons**: React Icons (Feather) for consistent iconography

## Currency Formatting

Components include built-in Indonesian Rupiah (IDR) formatting:
```typescript
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount);
};
```

## Error Handling

All components include comprehensive error handling:
- **API Error Display**: User-friendly error messages
- **Loading States**: Proper loading indicators
- **Validation**: Form validation and feedback
- **Toast Notifications**: Success/error notifications

## Usage Examples

### Basic Dashboard
```typescript
import React from 'react';
import { PaymentDashboard } from '@/components/payments';

const PaymentPage = () => {
  return (
    <div>
      <PaymentDashboard />
    </div>
  );
};
```

### Embedded Payment Form
```typescript
import React from 'react';
import { PaymentWithJournalForm } from '@/components/payments';

const CreatePaymentPage = () => {
  const handlePaymentCreated = (paymentId: string) => {
    // Redirect to payment details or show success
    console.log('Payment created:', paymentId);
  };

  return (
    <PaymentWithJournalForm
      onSubmit={handlePaymentCreated}
      initialData={{
        payment_type: 'RECEIVE',
        currency: 'IDR',
        payment_method: 'BANK_TRANSFER'
      }}
    />
  );
};
```

### Payment Details Modal
```typescript
import React from 'react';
import { Modal, ModalContent, ModalBody } from '@chakra-ui/react';
import { PaymentDetailsView } from '@/components/payments';

const PaymentModal = ({ isOpen, onClose, paymentId }) => {
  return (
    <Modal isOpen={isOpen} onClose={onClose} size="6xl">
      <ModalContent>
        <ModalBody>
          <PaymentDetailsView
            paymentId={paymentId}
            onEdit={(id) => console.log('Edit:', id)}
            onReverse={() => onClose()}
          />
        </ModalBody>
      </ModalContent>
    </Modal>
  );
};
```

## Development Notes

1. **Service Integration**: Ensure the payment service is properly configured
2. **Error Boundaries**: Consider wrapping components in error boundaries
3. **Testing**: Components are designed to be testable with proper prop interfaces
4. **Performance**: Large lists should implement pagination
5. **Accessibility**: Components follow ARIA guidelines through Chakra UI

## Future Enhancements

- **Pagination**: Add pagination for payment lists
- **Advanced Filters**: More filtering options
- **Bulk Operations**: Select and operate on multiple payments
- **Export Features**: Export payment and journal data
- **Audit Trail**: Show payment history and modifications
- **Real-time Updates**: WebSocket integration for live updates