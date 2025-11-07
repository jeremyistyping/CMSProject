const express = require('express');
const router = express.Router();
const purchaseReportController = require('../controllers/purchaseReportController');
const authMiddleware = require('../middleware/authMiddleware');

// Apply authentication middleware to all SSOT report routes
router.use(authMiddleware);

// Purchase Report endpoint (replaces deprecated vendor analysis)
router.get('/purchase-report', purchaseReportController.generatePurchaseReport);

// Legacy vendor analysis endpoint (deprecated)
router.get('/vendor-analysis', (req, res) => {
  res.status(400).json({
    success: false,
    error: 'This endpoint has been replaced by /api/ssot-reports/purchase-report',
    message: 'Vendor Analysis has been replaced with more credible Purchase Report',
    new_endpoints: {
      purchase_report: '/api/ssot-reports/purchase-report'
    },
    migration_guide: 'Use the new Purchase Report for accurate vendor and purchase analysis'
  });
});

// Health check endpoint
router.get('/health', (req, res) => {
  res.json({
    success: true,
    message: 'SSOT Reports API is healthy',
    version: '1.0.0',
    features: [
      'Purchase Report with SSOT Journal Integration',
      'Real-time vendor performance analysis',
      'Comprehensive payment tracking',
      'Outstanding payables monitoring'
    ]
  });
});

module.exports = router;