import React, { useState } from 'react';
import { Card, Form, Input, Select, Button, InputNumber, message, Spin, Statistic } from 'antd';
import { ThunderboltFilled, ClockCircleFilled } from '@ant-design/icons';
import axios from 'axios';

const { Option } = Select;

const UltraFastPaymentForm = ({ sales, cashBanks, onSuccess, onCancel }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [processingTime, setProcessingTime] = useState(null);
  const [lastPaymentId, setLastPaymentId] = useState(null);

  // Minimal validation - only essential fields
  const handleSubmit = async (values) => {
    const startTime = Date.now();
    setLoading(true);
    setProcessingTime(null);

    try {
      const response = await axios.post('/api/ultra-fast/payment', {
        sale_id: values.sale_id,
        amount: values.amount,
        cash_bank_id: values.cash_bank_id
      }, {
        timeout: 8000, // 8 second timeout
        headers: {
          'Content-Type': 'application/json'
        }
      });

      const endTime = Date.now();
      const clientProcessingTime = endTime - startTime;
      
      if (response.data.success) {
        setProcessingTime(response.data.processing_time);
        setLastPaymentId(response.data.payment_id);
        
        message.success({
          content: `⚡ Ultra-fast payment recorded in ${response.data.processing_time}! 
                    Client time: ${clientProcessingTime}ms`,
          duration: 5,
          style: { fontWeight: 'bold' }
        });

        form.resetFields();
        
        // Auto-call onSuccess if provided
        if (onSuccess) {
          setTimeout(() => onSuccess(response.data), 500);
        }
      } else {
        throw new Error(response.data.message || 'Payment failed');
      }
    } catch (error) {
      const endTime = Date.now();
      const clientProcessingTime = endTime - startTime;
      
      console.error('Ultra-fast payment error:', error);
      
      let errorMessage = 'Ultra-fast payment failed';
      if (error.response?.data?.message) {
        errorMessage = error.response.data.message;
      } else if (error.code === 'ECONNABORTED') {
        errorMessage = 'Payment timeout - server may be overloaded';
      } else if (error.message) {
        errorMessage = error.message;
      }

      message.error({
        content: `❌ ${errorMessage} (Client time: ${clientProcessingTime}ms)`,
        duration: 8
      });
    } finally {
      setLoading(false);
    }
  };

  // Get selected sale info for amount validation
  const selectedSale = sales?.find(sale => sale.id === form.getFieldValue('sale_id'));
  const maxAmount = selectedSale?.outstanding_amount || 0;

  return (
    <Card 
      title={
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <ThunderboltFilled style={{ color: '#faad14' }} />
          <span>⚡ Ultra-Fast Payment</span>
          {processingTime && (
            <Statistic 
              value={processingTime}
              prefix={<ClockCircleFilled />}
              suffix="processing"
              valueStyle={{ fontSize: '14px', color: '#52c41a' }}
            />
          )}
        </div>
      }
      size="small"
      style={{ maxWidth: 500 }}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        size="small"
      >
        <Form.Item
          name="sale_id"
          label="Sale"
          rules={[{ required: true, message: 'Please select a sale' }]}
        >
          <Select
            placeholder="Select sale to pay"
            showSearch
            optionFilterProp="children"
            disabled={loading}
            style={{ width: '100%' }}
          >
            {sales?.map(sale => (
              <Option key={sale.id} value={sale.id}>
                #{sale.invoice_number} - {sale.customer_name} 
                (Outstanding: ${sale.outstanding_amount?.toFixed(2)})
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          name="amount"
          label={`Payment Amount ${maxAmount ? `(Max: $${maxAmount.toFixed(2)})` : ''}`}
          rules={[
            { required: true, message: 'Please enter payment amount' },
            { type: 'number', min: 0.01, message: 'Amount must be greater than 0' },
            ...(maxAmount > 0 ? [{ 
              type: 'number', 
              max: maxAmount, 
              message: `Amount cannot exceed outstanding amount ($${maxAmount.toFixed(2)})` 
            }] : [])
          ]}
        >
          <InputNumber
            placeholder="Enter amount"
            min={0.01}
            max={maxAmount || undefined}
            step={0.01}
            precision={2}
            style={{ width: '100%' }}
            disabled={loading}
            formatter={value => `$ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
            parser={value => value.replace(/\$\s?|(,*)/g, '')}
          />
        </Form.Item>

        <Form.Item
          name="cash_bank_id"
          label="Cash/Bank Account"
          rules={[{ required: true, message: 'Please select cash/bank account' }]}
        >
          <Select
            placeholder="Select account"
            disabled={loading}
            style={{ width: '100%' }}
          >
            {cashBanks?.map(account => (
              <Option key={account.id} value={account.id}>
                {account.name} (Balance: ${account.balance?.toFixed(2)})
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item style={{ marginBottom: 0 }}>
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            {onCancel && (
              <Button 
                onClick={onCancel}
                disabled={loading}
              >
                Cancel
              </Button>
            )}
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              icon={<ThunderboltFilled />}
              style={{
                background: loading ? undefined : 'linear-gradient(135deg, #faad14 0%, #fa8c16 100%)',
                borderColor: loading ? undefined : '#faad14',
                fontWeight: 'bold'
              }}
            >
              {loading ? (
                <>
                  <Spin size="small" />
                  Processing...
                </>
              ) : (
                '⚡ Record Payment'
              )}
            </Button>
          </div>
        </Form.Item>

        {/* Processing time display */}
        {processingTime && (
          <div style={{ 
            marginTop: 16, 
            padding: 12, 
            background: '#f6ffed', 
            border: '1px solid #b7eb8f',
            borderRadius: 6,
            textAlign: 'center'
          }}>
            <div style={{ color: '#52c41a', fontWeight: 'bold' }}>
              ✅ Payment processed in {processingTime}
            </div>
            {lastPaymentId && (
              <div style={{ color: '#666', fontSize: '12px' }}>
                Payment ID: {lastPaymentId}
              </div>
            )}
          </div>
        )}
      </Form>

      {/* Performance notice */}
      <div style={{ 
        marginTop: 16, 
        padding: 8, 
        background: '#fff7e6', 
        border: '1px solid #ffd666',
        borderRadius: 4,
        fontSize: '12px',
        color: '#d46b08'
      }}>
        <ThunderboltFilled /> Ultra-Fast mode: Minimal validation, 5s timeout, async journal creation
      </div>
    </Card>
  );
};

export default UltraFastPaymentForm;