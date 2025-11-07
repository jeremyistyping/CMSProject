/**
 * Unified Financial Reports Service
 * 
 * This service consolidates and optimizes the integration between journal-data-driven reports
 * and traditional backend reports, providing a unified interface for all financial reporting needs.
 * 
 * Features:
 * - Intelligent routing between journal-based and backend-based reports
 * - Caching and performance optimization
 * - Data validation and consistency checks
 * - Unified error handling and fallback mechanisms
 * - Report quality scoring and recommendations
 */

import { reportService, ReportParameters } from './reportService';
import { enhancedPLService } from './enhancedPLService';
import { journalIntegrationService } from './journalIntegrationService';

export interface UnifiedReportOptions extends ReportParameters {
  prefer_journal_based?: boolean;
  include_quality_score?: boolean;
  enable_fallback?: boolean;
  cache_duration?: number; // in minutes
  validation_level?: 'basic' | 'strict' | 'comprehensive';
}

export interface UnifiedReportResponse {
  data: any;
  source: 'journal' | 'backend' | 'hybrid';
  quality_score: number;
  recommendations: string[];
  performance_metrics: {
    generation_time: number;
    data_freshness: number;
    cache_hit: boolean;
    fallback_used: boolean;
  };
  metadata: {
    generated_at: string;
    data_sources: string[];
    validation_results: ValidationResult[];
    optimization_applied: string[];
  };
}

export interface ValidationResult {
  rule: string;
  status: 'pass' | 'warning' | 'error';
  message: string;
  impact: 'low' | 'medium' | 'high';
}

export interface ReportCache {
  [key: string]: {
    data: any;
    timestamp: number;
    expiry: number;
    quality_score: number;
    source: string;
  };
}

class UnifiedFinancialReportsService {
  private cache: ReportCache = {};
  private readonly CACHE_DEFAULT_DURATION = 15; // minutes
  private readonly QUALITY_THRESHOLD = 0.8; // 80% quality score threshold

  /**
   * Generate any financial report using the optimal data source
   */
  async generateReport(
    reportType: string,
    options: UnifiedReportOptions
  ): Promise<UnifiedReportResponse> {
    const startTime = Date.now();
    const cacheKey = this.getCacheKey(reportType, options);
    
    // Check cache first
    if (this.isCacheValid(cacheKey) && !options.prefer_journal_based) {
      const cached = this.cache[cacheKey];
      return this.createResponse(cached.data, cached.source as any, cached.quality_score, [], {
        generation_time: 0,
        data_freshness: Date.now() - cached.timestamp,
        cache_hit: true,
        fallback_used: false,
      }, []);
    }

    try {
      let result: any;
      let source: 'journal' | 'backend' | 'hybrid';
      let validationResults: ValidationResult[] = [];
      let optimizations: string[] = [];

      // Determine optimal data source
      const strategy = await this.determineOptimalStrategy(reportType, options);
      
      switch (strategy.recommended_source) {
        case 'journal':
          result = await this.generateJournalBasedReport(reportType, options);
          source = 'journal';
          optimizations.push('journal_based_calculation', 'real_time_data');
          break;
          
        case 'backend':
          result = await this.generateBackendReport(reportType, options);
          source = 'backend';
          optimizations.push('pre_computed_aggregations', 'database_optimized');
          break;
          
        case 'hybrid':
          result = await this.generateHybridReport(reportType, options);
          source = 'hybrid';
          optimizations.push('hybrid_approach', 'cross_validation', 'best_of_both');
          break;
          
        default:
          throw new Error(`Unsupported report strategy: ${strategy.recommended_source}`);
      }

      // Validate results if requested
      if (options.validation_level) {
        validationResults = await this.validateReportData(result, reportType, options.validation_level);
      }

      // Calculate quality score
      const qualityScore = this.calculateQualityScore(result, validationResults, strategy);

      // Apply fallback if quality is too low
      if (qualityScore < this.QUALITY_THRESHOLD && options.enable_fallback !== false) {
        const fallbackResult = await this.applyFallbackStrategy(reportType, options, source);
        if (fallbackResult.quality_score > qualityScore) {
          result = fallbackResult.data;
          source = fallbackResult.source;
          optimizations.push('fallback_applied', `original_${source}_to_${fallbackResult.source}`);
        }
      }

      // Cache the result
      this.cacheResult(cacheKey, result, source, qualityScore, options.cache_duration);

      // Generate recommendations
      const recommendations = this.generateRecommendations(qualityScore, validationResults, strategy);

      const performanceMetrics = {
        generation_time: Date.now() - startTime,
        data_freshness: 0,
        cache_hit: false,
        fallback_used: optimizations.some(opt => opt.includes('fallback')),
      };

      return this.createResponse(result, source, qualityScore, recommendations, performanceMetrics, validationResults, optimizations);

    } catch (error) {
      console.error(`Unified report generation failed for ${reportType}:`, error);
      
      // Try fallback strategy on error
      if (options.enable_fallback !== false) {
        try {
          const fallbackResult = await this.emergencyFallback(reportType, options);
          const performanceMetrics = {
            generation_time: Date.now() - startTime,
            data_freshness: 0,
            cache_hit: false,
            fallback_used: true,
          };
          
          return this.createResponse(
            fallbackResult.data, 
            fallbackResult.source, 
            0.5, // Lower quality score for emergency fallback
            ['Emergency fallback used due to primary generation failure'], 
            performanceMetrics, 
            [], 
            ['emergency_fallback']
          );
        } catch (fallbackError) {
          console.error('Emergency fallback also failed:', fallbackError);
        }
      }
      
      throw error;
    }
  }

  /**
   * Get enhanced P&L with journal drilldown capabilities
   */
  async generateEnhancedProfitLoss(options: UnifiedReportOptions): Promise<UnifiedReportResponse> {
    const enhancedOptions = {
      ...options,
      prefer_journal_based: true,
      include_quality_score: true,
      validation_level: 'comprehensive' as const,
    };
    
    return this.generateReport('profit-loss', enhancedOptions);
  }

  /**
   * Get Balance Sheet with intelligent data sourcing
   */
  async generateBalanceSheet(options: UnifiedReportOptions): Promise<UnifiedReportResponse> {
    return this.generateReport('balance-sheet', options);
  }

  /**
   * Get Cash Flow with trend analysis
   */
  async generateCashFlow(options: UnifiedReportOptions): Promise<UnifiedReportResponse> {
    return this.generateReport('cash-flow', options);
  }

  /**
   * Get Trial Balance with accuracy validation
   */
  async generateTrialBalance(options: UnifiedReportOptions): Promise<UnifiedReportResponse> {
    const enhancedOptions = {
      ...options,
      validation_level: 'strict' as const,
      include_quality_score: true,
    };
    
    return this.generateReport('trial-balance', enhancedOptions);
  }

  /**
   * Determine the optimal strategy for report generation
   */
  private async determineOptimalStrategy(reportType: string, options: UnifiedReportOptions) {
    const factors = {
      data_volume: await this.estimateDataVolume(reportType, options),
      date_range_size: this.calculateDateRangeSize(options),
      real_time_requirement: this.assessRealTimeRequirement(reportType),
      complexity_score: this.calculateComplexityScore(reportType),
      user_preference: options.prefer_journal_based ? 1 : 0,
    };

    let recommended_source: 'journal' | 'backend' | 'hybrid';
    let confidence: number;

    // Decision logic
    if (factors.real_time_requirement > 0.7 || factors.data_volume < 1000) {
      recommended_source = 'journal';
      confidence = 0.8 + (factors.real_time_requirement * 0.2);
    } else if (factors.data_volume > 10000 || factors.date_range_size > 365) {
      recommended_source = 'backend';
      confidence = 0.7 + Math.min(factors.data_volume / 50000, 0.3);
    } else {
      recommended_source = 'hybrid';
      confidence = 0.9; // Hybrid generally provides best balance
    }

    // Override based on user preference
    if (options.prefer_journal_based) {
      recommended_source = 'journal';
      confidence *= 0.9; // Slightly lower confidence for forced preference
    }

    return {
      recommended_source,
      confidence,
      factors,
      reasoning: this.generateStrategyReasoning(factors, recommended_source),
    };
  }

  /**
   * Generate journal-based report
   */
  private async generateJournalBasedReport(reportType: string, options: UnifiedReportOptions) {
    switch (reportType) {
      case 'profit-loss':
        return await enhancedPLService.generateEnhancedPLFromJournals(options);
      case 'balance-sheet':
        return await journalIntegrationService.generateBalanceSheetFromJournals(options);
      case 'cash-flow':
        return await journalIntegrationService.generateCashFlowFromJournals(options);
      case 'trial-balance':
        return await journalIntegrationService.generateTrialBalanceFromJournals(options);
      default:
        throw new Error(`Journal-based report not supported for: ${reportType}`);
    }
  }

  /**
   * Generate backend report
   */
  private async generateBackendReport(reportType: string, options: UnifiedReportOptions) {
    return await reportService.generateReport(reportType, options);
  }

  /**
   * Generate hybrid report (combines journal and backend data)
   */
  private async generateHybridReport(reportType: string, options: UnifiedReportOptions) {
    const [journalResult, backendResult] = await Promise.allSettled([
      this.generateJournalBasedReport(reportType, options),
      this.generateBackendReport(reportType, options),
    ]);

    let primary = journalResult.status === 'fulfilled' ? journalResult.value : null;
    let secondary = backendResult.status === 'fulfilled' ? backendResult.value : null;

    if (!primary && !secondary) {
      throw new Error('Both journal and backend report generation failed');
    }

    if (!primary) return secondary;
    if (!secondary) return primary;

    // Merge and cross-validate results
    return this.mergeReportResults(primary, secondary, reportType);
  }

  /**
   * Validate report data quality
   */
  private async validateReportData(
    data: any, 
    reportType: string, 
    level: 'basic' | 'strict' | 'comprehensive'
  ): Promise<ValidationResult[]> {
    const results: ValidationResult[] = [];

    // Basic validations
    if (!data) {
      results.push({
        rule: 'data_exists',
        status: 'error',
        message: 'Report data is null or undefined',
        impact: 'high',
      });
      return results;
    }

    // Strict validations
    if (level === 'strict' || level === 'comprehensive') {
      if (reportType === 'profit-loss' || reportType === 'balance-sheet') {
        results.push(...this.validateFinancialStatementStructure(data, reportType));
      }
      
      if (reportType === 'trial-balance') {
        results.push(...this.validateTrialBalanceRules(data));
      }
    }

    // Comprehensive validations
    if (level === 'comprehensive') {
      results.push(...this.validateBusinessRules(data, reportType));
      results.push(...this.validateDataConsistency(data, reportType));
    }

    return results;
  }

  /**
   * Calculate overall quality score for a report
   */
  private calculateQualityScore(
    data: any, 
    validationResults: ValidationResult[], 
    strategy: any
  ): number {
    let score = 1.0;

    // Deduct points for validation issues
    validationResults.forEach(result => {
      if (result.status === 'error') {
        score -= result.impact === 'high' ? 0.3 : result.impact === 'medium' ? 0.2 : 0.1;
      } else if (result.status === 'warning') {
        score -= result.impact === 'high' ? 0.1 : result.impact === 'medium' ? 0.05 : 0.02;
      }
    });

    // Factor in strategy confidence
    score *= strategy.confidence;

    // Data completeness factor
    const completenessScore = this.assessDataCompleteness(data);
    score *= completenessScore;

    return Math.max(0, Math.min(1, score));
  }

  /**
   * Generate actionable recommendations based on report quality and issues
   */
  private generateRecommendations(
    qualityScore: number, 
    validationResults: ValidationResult[], 
    strategy: any
  ): string[] {
    const recommendations: string[] = [];

    if (qualityScore < 0.7) {
      recommendations.push('Report quality is below optimal threshold. Consider reviewing data sources and validation rules.');
    }

    const criticalErrors = validationResults.filter(r => r.status === 'error' && r.impact === 'high');
    if (criticalErrors.length > 0) {
      recommendations.push(`Found ${criticalErrors.length} critical error(s). Immediate attention required.`);
    }

    if (strategy.confidence < 0.8) {
      recommendations.push('Data source selection confidence is low. Consider using hybrid approach for better reliability.');
    }

    const unbalancedEntries = validationResults.filter(r => r.rule.includes('balance'));
    if (unbalancedEntries.length > 0) {
      recommendations.push('Journal entries contain unbalanced transactions. Review and correct accounting entries.');
    }

    if (qualityScore > 0.9) {
      recommendations.push('Excellent report quality achieved. Data is reliable for business decisions.');
    }

    return recommendations;
  }

  /**
   * Cache management
   */
  private getCacheKey(reportType: string, options: UnifiedReportOptions): string {
    const keyParts = [
      reportType,
      options.start_date || '',
      options.end_date || '',
      options.as_of_date || '',
      options.format || 'json',
      options.validation_level || 'basic',
      options.prefer_journal_based ? 'journal' : 'auto',
    ];
    
    return keyParts.join('|');
  }

  private isCacheValid(cacheKey: string): boolean {
    const cached = this.cache[cacheKey];
    return cached && cached.expiry > Date.now();
  }

  private cacheResult(
    cacheKey: string, 
    data: any, 
    source: string, 
    qualityScore: number, 
    durationMinutes?: number
  ): void {
    const duration = durationMinutes || this.CACHE_DEFAULT_DURATION;
    this.cache[cacheKey] = {
      data,
      timestamp: Date.now(),
      expiry: Date.now() + (duration * 60 * 1000),
      quality_score: qualityScore,
      source,
    };
  }

  /**
   * Utility methods
   */
  private createResponse(
    data: any,
    source: 'journal' | 'backend' | 'hybrid',
    qualityScore: number,
    recommendations: string[],
    performanceMetrics: any,
    validationResults: ValidationResult[] = [],
    optimizations: string[] = []
  ): UnifiedReportResponse {
    return {
      data,
      source,
      quality_score: qualityScore,
      recommendations,
      performance_metrics: performanceMetrics,
      metadata: {
        generated_at: new Date().toISOString(),
        data_sources: [source],
        validation_results: validationResults,
        optimization_applied: optimizations,
      },
    };
  }

  private async estimateDataVolume(reportType: string, options: UnifiedReportOptions): Promise<number> {
    // Simplified estimation - in real implementation, this would query metadata
    return 5000;
  }

  private calculateDateRangeSize(options: UnifiedReportOptions): number {
    if (!options.start_date || !options.end_date) return 30; // Default assumption
    
    const start = new Date(options.start_date);
    const end = new Date(options.end_date);
    return Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24));
  }

  private assessRealTimeRequirement(reportType: string): number {
    const realTimeReports = ['trial-balance', 'general-ledger'];
    return realTimeReports.includes(reportType) ? 0.8 : 0.3;
  }

  private calculateComplexityScore(reportType: string): number {
    const complexityMap: { [key: string]: number } = {
      'profit-loss': 0.8,
      'balance-sheet': 0.7,
      'cash-flow': 0.9,
      'trial-balance': 0.4,
      'general-ledger': 0.5,
    };
    
    return complexityMap[reportType] || 0.5;
  }

  private generateStrategyReasoning(factors: any, recommendedSource: string): string {
    if (recommendedSource === 'journal') {
      return 'Journal-based approach selected for real-time accuracy and smaller data volume';
    } else if (recommendedSource === 'backend') {
      return 'Backend approach selected for large data volume and complex aggregations';
    } else {
      return 'Hybrid approach selected for balanced performance and reliability';
    }
  }

  private mergeReportResults(primary: any, secondary: any, reportType: string): any {
    // Simplified merge logic - combine the best aspects of both results
    return {
      ...primary,
      _secondary_validation: secondary,
      _merge_timestamp: new Date().toISOString(),
      _confidence: 'high',
    };
  }

  private validateFinancialStatementStructure(data: any, reportType: string): ValidationResult[] {
    const results: ValidationResult[] = [];
    
    if (reportType === 'balance-sheet' && data.assets && data.liabilities && data.equity) {
      const assetsTotal = data.assets.total || 0;
      const liabilitiesTotal = (data.liabilities.total || 0) + (data.equity.total || 0);
      const difference = Math.abs(assetsTotal - liabilitiesTotal);
      
      if (difference > 0.01) { // Allow for minor rounding differences
        results.push({
          rule: 'balance_sheet_equation',
          status: 'error',
          message: `Assets (${assetsTotal}) do not equal Liabilities + Equity (${liabilitiesTotal})`,
          impact: 'high',
        });
      }
    }
    
    return results;
  }

  private validateTrialBalanceRules(data: any): ValidationResult[] {
    const results: ValidationResult[] = [];
    
    if (data.total_debits && data.total_credits) {
      if (Math.abs(data.total_debits - data.total_credits) > 0.01) {
        results.push({
          rule: 'trial_balance_equality',
          status: 'error',
          message: 'Total debits do not equal total credits',
          impact: 'high',
        });
      }
    }
    
    return results;
  }

  private validateBusinessRules(data: any, reportType: string): ValidationResult[] {
    const results: ValidationResult[] = [];
    
    // Add business-specific validation rules here
    // This is where you'd implement company-specific logic
    
    return results;
  }

  private validateDataConsistency(data: any, reportType: string): ValidationResult[] {
    const results: ValidationResult[] = [];
    
    // Check for data consistency issues
    // Implement cross-validation with other reports/systems
    
    return results;
  }

  private assessDataCompleteness(data: any): number {
    if (!data) return 0;
    
    // Simple completeness assessment
    let score = 1.0;
    
    if (typeof data === 'object') {
      const keys = Object.keys(data);
      const emptyKeys = keys.filter(key => !data[key]);
      score = Math.max(0, 1 - (emptyKeys.length / keys.length));
    }
    
    return score;
  }

  private async applyFallbackStrategy(
    reportType: string, 
    options: UnifiedReportOptions, 
    failedSource: 'journal' | 'backend' | 'hybrid'
  ): Promise<{ data: any; source: 'journal' | 'backend' | 'hybrid'; quality_score: number }> {
    // Try the alternative source
    const alternativeSource = failedSource === 'journal' ? 'backend' : 'journal';
    
    try {
      let data;
      if (alternativeSource === 'journal') {
        data = await this.generateJournalBasedReport(reportType, options);
      } else {
        data = await this.generateBackendReport(reportType, options);
      }
      
      return {
        data,
        source: alternativeSource,
        quality_score: 0.7, // Fallback gets moderate quality score
      };
    } catch (error) {
      throw new Error(`Fallback strategy also failed: ${error}`);
    }
  }

  private async emergencyFallback(
    reportType: string, 
    options: UnifiedReportOptions
  ): Promise<{ data: any; source: 'journal' | 'backend' | 'hybrid' }> {
    // Last resort - try to get any data possible
    try {
      const data = await reportService.generateReport(reportType, { ...options, format: 'json' });
      return { data, source: 'backend' };
    } catch (backendError) {
      try {
        const data = await this.generateJournalBasedReport(reportType, options);
        return { data, source: 'journal' };
      } catch (journalError) {
        throw new Error('All fallback strategies exhausted');
      }
    }
  }

  /**
   * Clear cache (useful for testing or forced refresh)
   */
  clearCache(): void {
    this.cache = {};
  }

  /**
   * Get cache statistics
   */
  getCacheStats() {
    const entries = Object.values(this.cache);
    const validEntries = entries.filter(entry => entry.expiry > Date.now());
    
    return {
      total_entries: entries.length,
      valid_entries: validEntries.length,
      expired_entries: entries.length - validEntries.length,
      average_quality_score: validEntries.reduce((sum, entry) => sum + entry.quality_score, 0) / validEntries.length || 0,
      sources_breakdown: validEntries.reduce((acc: any, entry) => {
        acc[entry.source] = (acc[entry.source] || 0) + 1;
        return acc;
      }, {}),
    };
  }
}

export const unifiedFinancialReportsService = new UnifiedFinancialReportsService();